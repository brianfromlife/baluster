package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

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
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

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
