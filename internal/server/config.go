package server

import (
	"fmt"
	"os"
	"time"
)

// Config holds server configuration
type Config struct {
	Port               string
	CosmosEndpoint     string
	CosmosKey          string
	CosmosDatabase     string
	JWTSecret          string
	JWTExpiration      time.Duration
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		CosmosEndpoint:     getEnv("COSMOS_ENDPOINT", ""),
		CosmosKey:          getEnv("COSMOS_KEY", ""),
		CosmosDatabase:     getEnv("COSMOS_DATABASE", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiration:      parseDuration(getEnv("JWT_EXPIRATION", "24h")),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:5173/auth/callback"),
	}

	// Validate required Cosmos DB environment variables
	if cfg.CosmosEndpoint == "" {
		return nil, fmt.Errorf("COSMOS_ENDPOINT environment variable is required")
	}
	if cfg.CosmosKey == "" {
		return nil, fmt.Errorf("COSMOS_KEY environment variable is required")
	}
	if cfg.CosmosDatabase == "" {
		return nil, fmt.Errorf("COSMOS_DATABASE environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
