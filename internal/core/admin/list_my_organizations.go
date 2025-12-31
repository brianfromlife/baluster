package admin

import (
	"context"

	"github.com/brianfromlife/baluster/internal/types"
)

type OrganizationMemberLister interface {
	ListByMemberID(ctx context.Context, userID string) ([]*types.Organization, error)
}

