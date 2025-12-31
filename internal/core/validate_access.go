package core

import (
	"context"

	"github.com/brianfromlife/baluster/internal/types"
)

type ServiceKeyTokenFinder interface {
	FindByTokenValue(ctx context.Context, tokenValue string) (*types.ServiceKey, error)
	FindByTokenValueInOrg(ctx context.Context, organizationID, tokenValue string) (*types.ServiceKey, error)
}

type ValidateAccessInput struct {
	Token           string
	ApplicationName string
	OrganizationID  string
}

type ValidateAccessOutput struct {
	Valid       bool
	Permissions []string
}

// ValidateAccess validates a service key for a specific application
// and returns the permissions it has for that application
func ValidateAccess(ctx context.Context, serviceKeyRepo ServiceKeyTokenFinder, input *ValidateAccessInput) (*ValidateAccessOutput, error) {
	if input.OrganizationID == "" {
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	serviceKey, err := serviceKeyRepo.FindByTokenValueInOrg(ctx, input.OrganizationID, input.Token)
	if err != nil {
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	if serviceKey.IsExpired() {
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	access := serviceKey.HasAccessToApplication(input.ApplicationName)
	if access == nil {
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	return &ValidateAccessOutput{
		Valid:       true,
		Permissions: access.Permissions,
	}, nil
}
