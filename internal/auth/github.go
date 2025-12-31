package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubConfig holds GitHub OAuth configuration
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// NewGitHubOAuth creates a new GitHub OAuth config
func NewGitHubOAuth(cfg GitHubConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint:     github.Endpoint,
	}
}

// GenerateState generates a random state string for OAuth
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGitHubUser fetches the authenticated user from GitHub
func GetGitHubUser(ctx context.Context, client *http.Client) (*GitHubUser, error) {
	logger := slog.Default()

	logger.Info("creating request to GitHub API")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		logger.Error("failed to create GitHub API request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	logger.Info("sending request to GitHub API")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send request to GitHub API", "error", err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	logger.Info("GitHub API response received", "status_code", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("GitHub API returned error",
			"status_code", resp.StatusCode,
			"body", string(body),
		)
		return nil, fmt.Errorf("github API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	logger.Info("decoding GitHub user response")
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		logger.Error("failed to decode GitHub user response", "error", err)
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	logger.Info("GitHub user decoded successfully",
		"id", user.ID,
		"login", user.Login,
		"email", user.Email,
	)

	return &user, nil
}

// StateCache provides thread-safe in-memory caching for OAuth state tokens
type StateCache struct {
	cache *Cache[bool]
}

// NewStateCache creates a new state cache
func NewStateCache() *StateCache {
	return &StateCache{
		cache: NewCache[bool](),
	}
}

// Set stores a state with expiration
func (c *StateCache) Set(state string, expiresAt time.Time) {
	c.cache.Set(state, true, expiresAt)
}

// Get checks if a state exists and is not expired
func (c *StateCache) Get(state string) bool {
	_, found := c.cache.Get(state)
	return found
}

// Delete removes a state
func (c *StateCache) Delete(state string) {
	c.cache.Delete(state)
}

// GetAndDelete atomically validates and removes a state
func (c *StateCache) GetAndDelete(state string) bool {
	_, found := c.cache.GetAndDelete(state)
	return found
}
