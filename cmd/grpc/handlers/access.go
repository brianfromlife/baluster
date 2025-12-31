package handlers

import (
	"context"

	"connectrpc.com/connect"

	"github.com/brianfromlife/baluster/internal/core"
	v1 "github.com/brianfromlife/baluster/internal/gen"
	"github.com/brianfromlife/baluster/internal/storage"
)

// AccessHandler implements the AccessService
type AccessHandler struct {
	serviceKeyRepo *storage.ServiceKeyRepository
}

// NewAccessHandler creates a new access handler
func NewAccessHandler(serviceKeyRepo *storage.ServiceKeyRepository) *AccessHandler {
	return &AccessHandler{
		serviceKeyRepo: serviceKeyRepo,
	}
}

// ValidateAccess validates a service key for a specific application and returns its permissions
func (h *AccessHandler) ValidateAccess(
	ctx context.Context,
	req *connect.Request[v1.ValidateAccessRequest],
) (*connect.Response[v1.ValidateAccessResponse], error) {
	// Extract organization ID from request message or metadata header (x-org-id)
	// Prefer the message field, but fall back to header for backwards compatibility
	orgID := req.Msg.OrganizationId
	if orgID == "" {
		orgID = req.Header().Get("x-org-id")
	}
	if orgID == "" {
		return connect.NewResponse(&v1.ValidateAccessResponse{
			Valid: false,
		}), connect.NewError(connect.CodeInvalidArgument, nil)
	}

	input := &core.ValidateAccessInput{
		Token:           req.Msg.Token,
		ApplicationName: req.Msg.ApplicationName,
		OrganizationID:  orgID,
	}

	output, err := core.ValidateAccess(ctx, h.serviceKeyRepo, input)
	if err != nil {
		return connect.NewResponse(&v1.ValidateAccessResponse{
			Valid: false,
		}), nil
	}

	return connect.NewResponse(&v1.ValidateAccessResponse{
		Valid:       output.Valid,
		Permissions: output.Permissions,
	}), nil
}
