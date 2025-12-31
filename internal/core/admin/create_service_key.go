package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyCreator interface {
	Create(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error
	CountByOrganization(ctx context.Context, organizationID string) (int, error)
}

// CreateServiceKeyInput represents the input for creating a service key
type CreateServiceKeyInput struct {
	Name         string
	Applications []types.ApplicationAccess
	ExpiresAt    *time.Time
}

// CreateServiceKeyOutput represents the output from creating a service key
type CreateServiceKeyOutput struct {
	ServiceKey *types.ServiceKey
	TokenValue string // The original unhashed token value (only returned on creation)
}

// CreateServiceKey creates a new service key
func CreateServiceKey(ctx context.Context, repo ServiceKeyCreator, input *CreateServiceKeyInput) (*CreateServiceKeyOutput, error) {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	// Check service key limit (max 50)
	count, err := repo.CountByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count existing service keys: %w", err)
	}
	if count >= 50 {
		return nil, ErrServiceKeyLimitExceeded
	}

	tokenValue := GenerateTokenValue()
	serviceKey := &types.ServiceKey{
		ID:                storage.GenerateID(),
		EntityType:        "service_key",
		OrganizationID:    orgID,
		Name:              input.Name,
		TokenValue:        tokenValue,
		Applications:      input.Applications,
		ExpiresAt:         input.ExpiresAt,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := repo.Create(ctx, serviceKey, userID, githubID, username); err != nil {
		return nil, err
	}

	return &CreateServiceKeyOutput{
		ServiceKey: serviceKey,
		TokenValue: tokenValue,
	}, nil
}
