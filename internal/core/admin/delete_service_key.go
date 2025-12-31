package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyDeleter interface {
	Get(ctx context.Context, organizationID, id string) (*types.ServiceKey, error)
	Delete(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error
}

// DeleteServiceKeyInput represents the input for deleting a service key
type DeleteServiceKeyInput struct {
	ID string
}

// DeleteServiceKey deletes a service key
func DeleteServiceKey(ctx context.Context, repo ServiceKeyDeleter, input *DeleteServiceKeyInput) error {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return fmt.Errorf("organization ID not found in context")
	}

	serviceKey, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return err
	}

	return repo.Delete(ctx, serviceKey, userID, githubID, username)
}
