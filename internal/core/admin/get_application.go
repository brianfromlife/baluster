package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApplicationGetter interface {
	Get(ctx context.Context, organizationID, id string) (*types.Application, error)
	GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error)
}

// GetApplicationInput represents the input for getting an application
type GetApplicationInput struct {
	ID string
}

// GetApplicationOutput represents the output from getting an application
type GetApplicationOutput struct {
	Application *types.Application
}

// GetApplication retrieves an application by ID
func GetApplication(ctx context.Context, repo ApplicationGetter, input *GetApplicationInput) (*GetApplicationOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	app, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetApplicationOutput{
		Application: app,
	}, nil
}

// GetApplicationHistoryInput represents the input for getting application audit history
type GetApplicationHistoryInput struct {
	ID string
}

// GetApplicationHistoryOutput represents the output from getting application audit history
type GetApplicationHistoryOutput struct {
	History []*types.AuditHistory
}

// GetApplicationHistory retrieves audit history for an application by ID
func GetApplicationHistory(ctx context.Context, repo ApplicationGetter, input *GetApplicationHistoryInput) (*GetApplicationHistoryOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	history, err := repo.GetHistory(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetApplicationHistoryOutput{
		History: history,
	}, nil
}
