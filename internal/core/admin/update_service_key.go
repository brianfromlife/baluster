package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyUpdater interface {
	Get(ctx context.Context, organizationID, id string) (*types.ServiceKey, error)
	Update(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error
}

// UpdateServiceKeyInput represents the input for updating a service key
type UpdateServiceKeyInput struct {
	ID          string
	Name        string
	Applications []types.ApplicationAccess
	ExpiresAt   *time.Time
}

// UpdateServiceKeyOutput represents the output from updating a service key
type UpdateServiceKeyOutput struct {
	ServiceKey *types.ServiceKey
}

// UpdateServiceKey updates a service key
func UpdateServiceKey(ctx context.Context, repo ServiceKeyUpdater, input *UpdateServiceKeyInput) (*UpdateServiceKeyOutput, error) {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	serviceKey, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	serviceKey.Name = input.Name
	serviceKey.Applications = input.Applications
	if input.ExpiresAt != nil {
		serviceKey.ExpiresAt = input.ExpiresAt
	}
	serviceKey.UpdatedAt = time.Now()

	if err := repo.Update(ctx, serviceKey, userID, githubID, username); err != nil {
		return nil, err
	}

	return &UpdateServiceKeyOutput{
		ServiceKey: serviceKey,
	}, nil
}
