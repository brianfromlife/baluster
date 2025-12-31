package core

import (
	"context"
	"log/slog"

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
	logger := slog.Default().With(
		"function", "ValidateAccess",
		"application_name", input.ApplicationName,
		"organization_id", input.OrganizationID,
	)
	logger.Info("validating access", "token_length", len(input.Token))

	if input.OrganizationID == "" {
		logger.Warn("organization ID is required")
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	serviceKey, err := serviceKeyRepo.FindByTokenValueInOrg(ctx, input.OrganizationID, input.Token)
	if err != nil {
		logger.Warn("failed to find service key by token value in organization", "error", err)
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	logger = logger.With(
		"service_key_id", serviceKey.ID,
		"organization_id", serviceKey.OrganizationID,
	)

	logger.Info("service key found",
		"applications_count", len(serviceKey.Applications),
		"applications", serviceKey.Applications,
		"expires_at", serviceKey.ExpiresAt,
	)

	if serviceKey.IsExpired() {
		logger.Warn("service key is expired", "expires_at", serviceKey.ExpiresAt)
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	logger.Debug("checking access to application",
		"requested_application_name", input.ApplicationName,
		"available_applications", serviceKey.Applications,
	)

	access := serviceKey.HasAccessToApplication(input.ApplicationName)
	if access == nil {
		logger.Warn("service key does not have access to application",
			"requested_application_name", input.ApplicationName,
			"available_application_names", getApplicationNames(serviceKey.Applications),
		)
		return &ValidateAccessOutput{
			Valid: false,
		}, nil
	}

	logger = logger.With(
		"application_id", access.ApplicationID,
		"permissions", access.Permissions,
	)

	var expiresAtUnix int64
	if serviceKey.ExpiresAt != nil {
		expiresAtUnix = serviceKey.ExpiresAt.Unix()
		logger = logger.With("expires_at", expiresAtUnix)
	}

	logger.Info("access validated successfully")

	return &ValidateAccessOutput{
		Valid:       true,
		Permissions: access.Permissions,
	}, nil
}

// getApplicationNames extracts application names from ApplicationAccess slice for logging
func getApplicationNames(apps []types.ApplicationAccess) []string {
	names := make([]string, len(apps))
	for i, app := range apps {
		names[i] = app.ApplicationName
	}
	return names
}
