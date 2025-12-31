package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"connectrpc.com/connect"

	"github.com/brianfromlife/baluster/cmd/grpc/handlers"
	"github.com/brianfromlife/baluster/internal/auth"
	balusterv1connect "github.com/brianfromlife/baluster/internal/gen/balusterv1connect"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/brianfromlife/baluster/internal/server"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env.local")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := server.LoadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	cosmosClient, err := storage.NewClient(ctx, storage.Config{
		Endpoint: cfg.CosmosEndpoint,
		Key:      cfg.CosmosKey,
		Database: cfg.CosmosDatabase,
	}, logger)
	if err != nil {
		logger.Error("failed to initialize Cosmos DB client", "error", err)
		os.Exit(1)
	}

	// Initialize repositories (only what we need for ValidateAccess)
	serviceKeyRepo, err := storage.NewServiceKeyRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize service key repository", "error", err)
		os.Exit(1)
	}

	apiKeyRepo, err := storage.NewApiKeyRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize API key repository", "error", err)
		os.Exit(1)
	}

	apiKeyValidator := auth.NewApiKeyValidator(apiKeyRepo)
	accessHandler := handlers.NewAccessHandler(serviceKeyRepo)

	mux := http.NewServeMux()

	accessPath, accessHandlerHTTP := balusterv1connect.NewAccessServiceHandler(
		accessHandler,
		connect.WithInterceptors(auth.ApiKeyAuthInterceptor(apiKeyValidator)),
	)

	mux.Handle(accessPath, accessHandlerHTTP)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      h2c.NewHandler(httputil.CORS()(mux), &http2.Server{}),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting gRPC server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
}
