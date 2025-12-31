package admin

import (
	"context"
	"fmt"
	"log/slog"
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
	logger := slog.Default().With("state", input.State)

	logger.Info("validating OAuth state")
	valid := stateCache.GetAndDelete(input.State)
	if !valid {
		logger.Error("state validation failed", "state", input.State)
		return nil, fmt.Errorf("invalid state")
	}

	logger.Info("state validated successfully")

	logger.Info("exchanging OAuth code for token")
	token, err := githubOAuth.Exchange(ctx, input.Code)
	if err != nil {
		logger.Error("OAuth token exchange failed", "error", err)
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	logger.Info("OAuth token exchange successful")

	logger.Info("fetching GitHub user information")
	client := githubOAuth.Client(ctx, token)
	githubUser, err := auth.GetGitHubUser(ctx, client)
	if err != nil {
		logger.Error("failed to fetch GitHub user", "error", err)
		return nil, fmt.Errorf("failed to fetch GitHub user: %w", err)
	}
	logger.Info("GitHub user fetched",
		"github_id", githubUser.ID,
		"username", githubUser.Login,
		"email", githubUser.Email,
	)

	user := &types.User{
		ID:        storage.GenerateID(),
		GitHubID:  fmt.Sprintf("%d", githubUser.ID),
		Username:  githubUser.Login,
		AvatarURL: githubUser.AvatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logger.Info("checking for existing user by GitHub ID", "github_id", user.GitHubID)
	existingUser, err := userRepo.GetByGitHubID(ctx, user.GitHubID)
	if err == nil {
		logger.Info("existing user found", "user_id", existingUser.ID)
		user.ID = existingUser.ID
		user.CreatedAt = existingUser.CreatedAt
	} else {
		logger.Info("no existing user found, creating new user")
	}

	user.UpdatedAt = time.Now()
	logger.Info("saving user to repository", "user_id", user.ID)
	if err := userRepo.CreateOrUpdate(ctx, user); err != nil {
		logger.Error("failed to create or update user", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	logger.Info("user saved successfully", "user_id", user.ID)

	logger.Info("generating JWT token", "user_id", user.ID)
	jwtToken, err := auth.GenerateToken(jwtConfig, user.ID, user.GitHubID, user.Username)
	if err != nil {
		logger.Error("failed to generate JWT token", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	logger.Info("JWT token generated successfully", "user_id", user.ID)

	return &GitHubOAuthCallbackOutput{
		SessionToken: jwtToken,
		User:         user,
	}, nil
}
