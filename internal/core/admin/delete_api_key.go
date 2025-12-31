package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApiKeyDeleter interface {
	Get(ctx context.Context, organizationID, id string) (*types.ApiKey, error)
	Delete(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error
}

// DeleteApiKeyInput represents the input for deleting an API key
type DeleteApiKeyInput struct {
	ID string
}

// DeleteApiKey deletes an API key
func DeleteApiKey(ctx context.Context, repo ApiKeyDeleter, input *DeleteApiKeyInput) error {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return fmt.Errorf("organization ID not found in context")
	}

	token, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return err
	}

	return repo.Delete(ctx, token, userID, githubID, username)
}
