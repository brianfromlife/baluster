package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/brianfromlife/baluster/cmd/rest/handlers"
	"github.com/brianfromlife/baluster/internal/auth"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/brianfromlife/baluster/internal/server"
	"github.com/brianfromlife/baluster/internal/storage"
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

	// Cosmos
	ctx := context.Background()
	cosmosClient, err := storage.NewClient(ctx, storage.Config{
		Endpoint: cfg.CosmosEndpoint,
		Key:      cfg.CosmosKey,
		Database: cfg.CosmosDatabase,
	})
	if err != nil {
		logger.Error("failed to initialize Cosmos DB client", "error", err)
		os.Exit(1)
	}

	// Repositories
	orgRepo, err := storage.NewOrganizationRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize organization repository", "error", err)
		os.Exit(1)
	}

	appRepo, err := storage.NewApplicationRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize application repository", "error", err)
		os.Exit(1)
	}

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

	userRepo, err := storage.NewUserRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize user repository", "error", err)
		os.Exit(1)
	}

	orgMemberRepo, err := storage.NewOrganizationMemberRepository(cosmosClient)
	if err != nil {
		logger.Error("failed to initialize organization member repository", "error", err)
		os.Exit(1)
	}

	jwtConfig := auth.JWTConfig{
		Secret:     cfg.JWTSecret,
		Expiration: cfg.JWTExpiration,
	}

	githubConfig := auth.GitHubConfig{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURL,
		Scopes:       []string{"read:user"},
	}

	githubOAuth := auth.NewGitHubOAuth(githubConfig)
	stateCache := auth.NewStateCache()
	membershipCache := auth.NewMembershipCache(5 * time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httputil.CORS(httputil.CORSOptions{
		AllowedHeaders: []string{"x-org-id"},
		ExposeHeaders:  []string{"Link", auth.TokenRefreshHeader},
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Initialize validators
	apiKeyValidator := auth.NewApiKeyValidator(apiKeyRepo)

	authHandler := handlers.NewAuthHandler(githubOAuth, stateCache, jwtConfig, userRepo, orgRepo)
	r.Route("/api/auth", authHandler.RegisterRoutes)

	// Admin API routes (GitHub OAuth)
	r.Route("/admin/v1", func(r chi.Router) {
		r.Use(auth.JWTAuthMiddleware(jwtConfig, userRepo))

		// User routes (no org context needed)
		r.Get("/me", authHandler.GetCurrentUser)

		// Organization routes (no org context needed for creation)
		r.Post("/organizations", handlers.CreateOrganization(orgRepo, orgMemberRepo))

		// Routes that require organization membership
		r.Group(func(r chi.Router) {
			r.Use(auth.OrganizationMembershipMiddleware(orgMemberRepo, membershipCache))

			// Organization-scoped list routes (these URLs still have org_id for clarity, but middleware validates)
			r.Get("/organizations/{organization_id}/applications", handlers.ListApplications(appRepo))
			r.Get("/organizations/{organization_id}/service-keys", handlers.ListServiceKeys(serviceKeyRepo))
			r.Get("/organizations/{organization_id}/api-keys", handlers.ListApiKeys(apiKeyRepo))

			// Application routes
			r.Post("/applications", handlers.CreateApplication(appRepo))
			r.Get("/applications/{application_id}", handlers.GetApplication(appRepo))
			r.Get("/applications/{application_id}/history", handlers.GetApplicationHistory(appRepo))
			r.Put("/applications/{application_id}", handlers.UpdateApplication(appRepo))

			// Service key routes
			r.Post("/service-keys", handlers.CreateServiceKey(serviceKeyRepo))
			r.Get("/service-keys/{service_key_id}", handlers.GetServiceKey(serviceKeyRepo))
			r.Get("/service-keys/{service_key_id}/history", handlers.GetServiceKeyHistory(serviceKeyRepo))
			r.Put("/service-keys/{service_key_id}", handlers.UpdateServiceKey(serviceKeyRepo))
			r.Delete("/service-keys/{service_key_id}", handlers.DeleteServiceKey(serviceKeyRepo))

			// API key routes
			r.Post("/api-keys", handlers.CreateApiKey(apiKeyRepo))
			r.Get("/api-keys/{token_id}", handlers.GetApiKey(apiKeyRepo))
			r.Get("/api-keys/{token_id}/history", handlers.GetApiKeyHistory(apiKeyRepo))
			r.Put("/api-keys/{token_id}", handlers.UpdateApiKey(apiKeyRepo))
			r.Delete("/api-keys/{token_id}", handlers.DeleteApiKey(apiKeyRepo))
		})
	})

	// Service key validation
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.ApiKeyAuthMiddleware(apiKeyValidator))
		r.Post("/access", handlers.ValidateAccess(serviceKeyRepo))
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting REST server", "port", cfg.Port)
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
