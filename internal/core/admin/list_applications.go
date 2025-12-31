package admin

import (
	"context"
	"fmt"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/types"
)

type ApplicationLister interface {
	ListByOrganization(ctx context.Context, organizationID string) ([]*types.Application, error)
}

// ListApplicationsInput represents the input for listing applications
type ListApplicationsInput struct {
}

// ListApplicationsOutput represents the output from listing applications
type ListApplicationsOutput struct {
	Applications []*types.Application
}

// ListApplications lists applications for an organization
func ListApplications(ctx context.Context, repo ApplicationLister, input *ListApplicationsInput) (*ListApplicationsOutput, error) {
	orgID, ok := auth.GetOrganizationID(ctx)
	if !ok {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	apps, err := repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &ListApplicationsOutput{
		Applications: apps,
	}, nil
}
