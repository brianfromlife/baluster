package admin

import (
	"context"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"golang.org/x/oauth2"
)

// GitHubOAuthInput represents the input for initiating GitHub OAuth
type GitHubOAuthInput struct {
	RedirectURI string
}

// GitHubOAuthOutput represents the output from initiating GitHub OAuth
type GitHubOAuthOutput struct {
	AuthURL string
	State   string
}

// GitHubOAuth initiates GitHub OAuth flow
func GitHubOAuth(ctx context.Context, githubOAuth *oauth2.Config, stateCache *auth.StateCache, input *GitHubOAuthInput) (*GitHubOAuthOutput, error) {
	redirectURI := input.RedirectURI
	if redirectURI == "" {
		redirectURI = githubOAuth.RedirectURL
	}

	state, err := auth.GenerateState()
	if err != nil {
		return nil, err
	}

	// Store state with 10 minute expiration
	stateCache.Set(state, time.Now().Add(10*time.Minute))

	authURL := githubOAuth.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return &GitHubOAuthOutput{
		AuthURL: authURL,
		State:   state,
	}, nil
}
