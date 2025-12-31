package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyGetter interface {
	Get(ctx context.Context, organizationID, id string) (*types.ServiceKey, error)
	GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error)
}

// GetServiceKeyInput represents the input for getting a service key
type GetServiceKeyInput struct {
	ID string
}

// GetServiceKeyOutput represents the output from getting a service key
type GetServiceKeyOutput struct {
	ServiceKey *types.ServiceKey
}

// GetServiceKey retrieves a service key by ID (without token value)
func GetServiceKey(ctx context.Context, repo ServiceKeyGetter, input *GetServiceKeyInput) (*GetServiceKeyOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	serviceKey, err := repo.Get(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetServiceKeyOutput{
		ServiceKey: serviceKey,
	}, nil
}

// GetServiceKeyHistoryInput represents the input for getting service key audit history
type GetServiceKeyHistoryInput struct {
	ID string
}

// GetServiceKeyHistoryOutput represents the output from getting service key audit history
type GetServiceKeyHistoryOutput struct {
	History []*types.AuditHistory
}

// GetServiceKeyHistory retrieves audit history for a service key by ID
func GetServiceKeyHistory(ctx context.Context, repo ServiceKeyGetter, input *GetServiceKeyHistoryInput) (*GetServiceKeyHistoryOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	history, err := repo.GetHistory(ctx, orgID, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetServiceKeyHistoryOutput{
		History: history,
	}, nil
}
