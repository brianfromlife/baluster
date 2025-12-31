package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/storage"
	"github.com/brianfromlife/baluster/internal/types"
)

type OrganizationCreator interface {
	Create(ctx context.Context, org *types.Organization) error
}

type OrganizationMemberAdder interface {
	AddMember(ctx context.Context, orgID, userID string) error
}

type OrganizationWithMemberCreator interface {
	CreateOrganizationWithMember(ctx context.Context, org *types.Organization, userID string) error
}

type CreateOrganizationInput struct {
	Name string
}

type CreateOrganizationOutput struct {
	Organization *types.Organization
}

func CreateOrganization(ctx context.Context, repo OrganizationCreator, memberRepo OrganizationMemberAdder, input *CreateOrganizationInput) (*CreateOrganizationOutput, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUserInfoNotFound
	}

	org := &types.Organization{
		ID:        storage.GenerateID(),
		Name:      input.Name,
		MemberIDs: []string{}, // Empty - membership is now stored as organization_member records in organizations container
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Try to use the atomic CreateOrganizationWithMember if available
	if atomicCreator, ok := memberRepo.(OrganizationWithMemberCreator); ok {
		if err := atomicCreator.CreateOrganizationWithMember(ctx, org, userID); err != nil {
			return nil, err
		}
		return &CreateOrganizationOutput{
			Organization: org,
		}, nil
	}

	// Fallback to two-step process (less safe, but maintains backward compatibility)
	if err := repo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Add creator as member (organization_member record in organizations container)
	// This is required - if it fails, the organization creation should fail
	if memberRepo == nil {
		return nil, fmt.Errorf("member repository is required")
	}
	if err := memberRepo.AddMember(ctx, org.ID, userID); err != nil {
		return nil, fmt.Errorf("failed to add creator as organization member: %w", err)
	}

	return &CreateOrganizationOutput{
		Organization: org,
	}, nil
}
