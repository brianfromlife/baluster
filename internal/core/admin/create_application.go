package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApplicationCreator interface {
	Create(ctx context.Context, app *types.Application, userID, githubID, username string) error
	CountByOrganization(ctx context.Context, organizationID string) (int, error)
}

type CreateApplicationInput struct {
	Name        string
	Description string
	Permissions []string // e.g., ["user.delete", "order.cancel", "admin", "basic"]
}

type CreateApplicationOutput struct {
	Application *types.Application
}

func CreateApplication(ctx context.Context, repo ApplicationCreator, input *CreateApplicationInput) (*CreateApplicationOutput, error) {
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	// Check application limit (max 20)
	count, err := repo.CountByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count existing applications: %w", err)
	}
	if count >= 20 {
		return nil, ErrApplicationLimitExceeded
	}

	app := &types.Application{
		ID:                storage.GenerateID(),
		EntityType:        "application",
		OrganizationID:    orgID,
		Name:              input.Name,
		Description:       input.Description,
		Permissions:       input.Permissions,
		CreatedByUserID:   userID,
		CreatedByGitHubID: githubID,
		CreatedByUsername: username,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := repo.Create(ctx, app, userID, githubID, username); err != nil {
		return nil, err
	}

	return &CreateApplicationOutput{
		Application: app,
	}, nil
}
