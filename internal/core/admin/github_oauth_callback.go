package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/brianfromlife/baluster/internal/types"
	"golang.org/x/oauth2"
)

type UserStore interface {
	GetByGitHubID(ctx context.Context, githubID string) (*types.User, error)
	CreateOrUpdate(ctx context.Context, user *types.User) error
}

type GitHubOAuthCallbackInput struct {
	Code  string
	State string
}

type GitHubOAuthCallbackOutput struct {
	SessionToken string
	User         *types.User
}

func GitHubOAuthCallback(ctx context.Context, githubOAuth *oauth2.Config, stateCache *auth.StateCache, jwtConfig auth.JWTConfig, userRepo UserStore, input *GitHubOAuthCallbackInput) (*GitHubOAuthCallbackOutput, error) {
	valid := stateCache.GetAndDelete(input.State)
	if !valid {
		return nil, fmt.Errorf("invalid state")
	}

	token, err := githubOAuth.Exchange(ctx, input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := githubOAuth.Client(ctx, token)
	githubUser, err := auth.GetGitHubUser(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub user: %w", err)
	}

	user := &types.User{
		ID:        storage.GenerateID(),
		GitHubID:  fmt.Sprintf("%d", githubUser.ID),
		Username:  githubUser.Login,
		AvatarURL: githubUser.AvatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	existingUser, err := userRepo.GetByGitHubID(ctx, user.GitHubID)
	if err == nil {
		user.ID = existingUser.ID
		user.CreatedAt = existingUser.CreatedAt
	}

	user.UpdatedAt = time.Now()
	if err := userRepo.CreateOrUpdate(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	jwtToken, err := auth.GenerateToken(jwtConfig, user.ID, user.GitHubID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &GitHubOAuthCallbackOutput{
		SessionToken: jwtToken,
		User:         user,
	}, nil
}
