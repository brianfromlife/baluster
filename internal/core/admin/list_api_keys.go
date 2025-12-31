package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApiKeyLister interface {
	ListByOrganization(ctx context.Context, organizationID string) ([]*types.ApiKey, error)
}

// ListApiKeysInput represents the input for listing API keys
type ListApiKeysInput struct {
}

// ListApiKeysOutput represents the output from listing API keys
type ListApiKeysOutput struct {
	Tokens []*types.ApiKey
}

// ListApiKeys lists API keys for an organization
func ListApiKeys(ctx context.Context, repo ApiKeyLister, input *ListApiKeysInput) (*ListApiKeysOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	tokens, err := repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &ListApiKeysOutput{
		Tokens: tokens,
	}, nil
}
