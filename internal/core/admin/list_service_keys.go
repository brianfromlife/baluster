package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyLister interface {
	ListByOrganization(ctx context.Context, organizationID string) ([]*types.ServiceKey, error)
}

// ListServiceKeysInput represents the input for listing service keys
type ListServiceKeysInput struct {
}

// ListServiceKeysOutput represents the output from listing service keys
type ListServiceKeysOutput struct {
	ServiceKeys []*types.ServiceKey
}

// ListServiceKeys lists service keys for an organization
func ListServiceKeys(ctx context.Context, repo ServiceKeyLister, input *ListServiceKeysInput) (*ListServiceKeysOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	serviceKeys, err := repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &ListServiceKeysOutput{
		ServiceKeys: serviceKeys,
	}, nil
}
