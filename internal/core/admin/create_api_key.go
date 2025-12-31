package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApiKeyCreator interface {
	Create(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error
	CountByOrganization(ctx context.Context, organizationID string) (int, error)
}

type CreateApiKeyInput struct {
	ApplicationID string
	Name         string
	ExpiresAt    *time.Time
}

type CreateApiKeyOutput struct {
	Token      *types.ApiKey
	TokenValue string
}

// CreateApiKey creates a new API key
func CreateApiKey(ctx context.Context, repo ApiKeyCreator, input *CreateApiKeyInput) (*CreateApiKeyOutput, error) {
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	// Check API key limit (max 50)
	count, err := repo.CountByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count existing API keys: %w", err)
	}
	if count >= 50 {
		return nil, ErrApiKeyLimitExceeded
	}

	tokenValue := GenerateTokenValue()
	token := &types.ApiKey{
		ID:                storage.GenerateID(),
		EntityType:        "api_key",
		OrganizationID:    orgID,
		ApplicationID:     input.ApplicationID,
		Name:              input.Name,
		TokenValue:        tokenValue,
		ExpiresAt:         input.ExpiresAt,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := repo.Create(ctx, token, userID, githubID, username); err != nil {
		return nil, err
	}

	return &CreateApiKeyOutput{
		Token:      token,
		TokenValue: tokenValue,
	}, nil
}
