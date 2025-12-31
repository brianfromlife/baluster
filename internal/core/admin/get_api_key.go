package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApiKeyGetter interface {
	Get(ctx context.Context, organizationID, id string) (*types.ApiKey, error)
	GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error)
}

// GetApiKeyInput represents the input for getting an API key
type GetApiKeyInput struct {
	ID string
}

// GetApiKeyOutput represents the output from getting an API key
type GetApiKeyOutput struct {
	Token *types.ApiKey
}

// GetApiKey retrieves an API key by ID (without token value)
func GetApiKey(ctx context.Context, repo ApiKeyGetter, input *GetApiKeyInput) (*GetApiKeyOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	token, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetApiKeyOutput{
		Token: token,
	}, nil
}

// GetApiKeyHistoryInput represents the input for getting API key audit history
type GetApiKeyHistoryInput struct {
	ID string
}

// GetApiKeyHistoryOutput represents the output from getting API key audit history
type GetApiKeyHistoryOutput struct {
	History []*types.AuditHistory
}

// GetApiKeyHistory retrieves audit history for an API key by ID
func GetApiKeyHistory(ctx context.Context, repo ApiKeyGetter, input *GetApiKeyHistoryInput) (*GetApiKeyHistoryOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	history, err := repo.GetHistory(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetApiKeyHistoryOutput{
		History: history,
	}, nil
}
