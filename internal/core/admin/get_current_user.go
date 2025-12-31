package admin

import (
	"context"

	"github.com/brianfromlife/baluster/internal/types"
)

type UserGetter interface {
	GetByID(ctx context.Context, id string, githubID string) (*types.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*types.User, error)
}

type GetCurrentUserInput struct {
	UserID   string
	GithubID string
}

type GetCurrentUserOutput struct {
	User         *types.User
	Organization *types.Organization
}

// GetCurrentUser returns the current authenticated user with their organizations
func GetCurrentUser(ctx context.Context, userRepo UserGetter, orgRepo OrganizationMemberLister, input *GetCurrentUserInput) (*GetCurrentUserOutput, error) {
	// Use GetByGitHubID since we only have GitHubID
	user, err := userRepo.GetByID(ctx, input.UserID, input.GithubID)
	if err != nil {
		return nil, err
	}

	orgs, err := orgRepo.ListByMemberID(ctx, input.UserID)
	if err != nil {
		// If organization fetch fails, return user without organization
		return &GetCurrentUserOutput{
			User:         user,
			Organization: nil,
		}, nil
	}

	// Return first organization if available, otherwise nil
	var org *types.Organization
	if len(orgs) > 0 {
		org = orgs[0]
	}

	return &GetCurrentUserOutput{
		User:         user,
		Organization: org,
	}, nil
}
