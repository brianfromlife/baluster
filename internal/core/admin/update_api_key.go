package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApiKeyUpdater interface {
	Get(ctx context.Context, organizationID, id string) (*types.ApiKey, error)
	Update(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error
}

// UpdateApiKeyInput represents the input for updating an API key
type UpdateApiKeyInput struct {
	ID        string
	Name      string
	ExpiresAt *time.Time
}

// UpdateApiKeyOutput represents the output from updating an API key
type UpdateApiKeyOutput struct {
	Token *types.ApiKey
}

// UpdateApiKey updates an API key
func UpdateApiKey(ctx context.Context, repo ApiKeyUpdater, input *UpdateApiKeyInput) (*UpdateApiKeyOutput, error) {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	token, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	token.Name = input.Name
	if input.ExpiresAt != nil {
		token.ExpiresAt = input.ExpiresAt
	}
	token.UpdatedAt = time.Now()

	if err := repo.Update(ctx, token, userID, githubID, username); err != nil {
		return nil, err
	}

	return &UpdateApiKeyOutput{
		Token: token,
	}, nil
}
