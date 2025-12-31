package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApplicationUpdater interface {
	Get(ctx context.Context, organizationID, id string) (*types.Application, error)
	Update(ctx context.Context, app *types.Application, userID, githubID, username string) error
}

// UpdateApplicationInput represents the input for updating an application
type UpdateApplicationInput struct {
	ID          string
	Name        string
	Description string
	Permissions []string
}

// UpdateApplicationOutput represents the output from updating an application
type UpdateApplicationOutput struct {
	Application *types.Application
}

// UpdateApplication updates an application
func UpdateApplication(ctx context.Context, repo ApplicationUpdater, input *UpdateApplicationInput) (*UpdateApplicationOutput, error) {
	// Extract user information from context
	userID, githubID, username, ok := auth.GetUserInfo(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	app, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	app.Name = input.Name
	app.Description = input.Description
	app.Permissions = input.Permissions
	app.UpdatedAt = time.Now()

	if err := repo.Update(ctx, app, userID, githubID, username); err != nil {
		return nil, err
	}

	return &UpdateApplicationOutput{
		Application: app,
	}, nil
}
