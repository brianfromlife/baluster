package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brianfromlife/baluster/internal/core"
	"github.com/brianfromlife/baluster/internal/core/admin"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/brianfromlife/baluster/internal/types"
	"github.com/go-chi/chi/v5"
)

// ApplicationAccessRequest represents an application access entry in requests
type ApplicationAccessRequest struct {
	ApplicationID   string   `json:"application_id"`
	ApplicationName string   `json:"application_name"`
	Permissions     []string `json:"permissions"`
}

type CreateServiceKeyRequest struct {
	Name         string                     `json:"name"`
	Applications []ApplicationAccessRequest `json:"applications"`
	ExpiresAt    *time.Time                 `json:"expires_at"`
}

func (r CreateServiceKeyRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	return nil
}

type CreateServiceKeyResponse struct {
	ServiceKey     any    `json:"service_key"`
	TokenValue     string `json:"token_value"`
	OrganizationID string `json:"organization_id"`
}

// CreateServiceKey creates a new service key
func CreateServiceKey(serviceKeyRepo admin.ServiceKeyCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[CreateServiceKeyRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		// Convert request applications to types.ApplicationAccess
		var applications []types.ApplicationAccess
		for _, app := range req.Applications {
			applications = append(applications, types.ApplicationAccess{
				ApplicationID:   app.ApplicationID,
				ApplicationName: app.ApplicationName,
				Permissions:     app.Permissions,
			})
		}

		input := &admin.CreateServiceKeyInput{
			Name:         req.Name,
			Applications: applications,
			ExpiresAt:    req.ExpiresAt,
		}

		output, err := admin.CreateServiceKey(r.Context(), serviceKeyRepo, input)
		if err != nil {
			if errors.Is(err, admin.ErrUserInfoNotFound) {
				httputil.Error(w, http.StatusUnauthorized, err)
				return
			}
			if errors.Is(err, admin.ErrServiceKeyLimitExceeded) {
				httputil.Error(w, http.StatusForbidden, err)
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		responseServiceKey := output.ServiceKey
		responseServiceKey.TokenValue = ""

		httputil.Success(w, http.StatusCreated, CreateServiceKeyResponse{
			ServiceKey:     responseServiceKey,
			TokenValue:     output.TokenValue,
			OrganizationID: responseServiceKey.OrganizationID,
		})
	}
}

// GetServiceKey gets a service key by ID
func GetServiceKey(serviceKeyRepo admin.ServiceKeyGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceKeyID := chi.URLParam(r, "service_key_id")
		if serviceKeyID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service_key_id is required"))
			return
		}

		input := &admin.GetServiceKeyInput{
			ID: serviceKeyID,
		}

		output, err := admin.GetServiceKey(r.Context(), serviceKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		output.ServiceKey.TokenValue = ""

		httputil.Success(w, http.StatusOK, output.ServiceKey)
	}
}

// GetServiceKeyHistory gets audit history for a service key by ID
func GetServiceKeyHistory(serviceKeyRepo admin.ServiceKeyGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceKeyID := chi.URLParam(r, "service_key_id")
		if serviceKeyID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service_key_id is required"))
			return
		}

		input := &admin.GetServiceKeyHistoryInput{
			ID: serviceKeyID,
		}

		output, err := admin.GetServiceKeyHistory(r.Context(), serviceKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, map[string]any{
			"history": output.History,
		})
	}
}

// ListServiceKeys lists service keys for an organization
func ListServiceKeys(serviceKeyRepo admin.ServiceKeyLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := &admin.ListServiceKeysInput{}

		output, err := admin.ListServiceKeys(r.Context(), serviceKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear token values from response
		for _, serviceKey := range output.ServiceKeys {
			serviceKey.TokenValue = ""
		}

		httputil.Success(w, http.StatusOK, output.ServiceKeys)
	}
}

// UpdateServiceKeyRequest represents the HTTP request to update a service key
type UpdateServiceKeyRequest struct {
	Name         string                     `json:"name"`
	Applications []ApplicationAccessRequest `json:"applications"`
	ExpiresAt    *time.Time                 `json:"expires_at"`
}

// Validate validates the UpdateServiceKeyRequest
func (r UpdateServiceKeyRequest) Validate() error {
	if r.Name != "" && len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	return nil
}

// UpdateServiceKey updates a service key
func UpdateServiceKey(serviceKeyRepo admin.ServiceKeyUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceKeyID := chi.URLParam(r, "service_key_id")
		if serviceKeyID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service_key_id is required"))
			return
		}

		req, err := httputil.Decode[UpdateServiceKeyRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		var applications []types.ApplicationAccess
		for _, app := range req.Applications {
			applications = append(applications, types.ApplicationAccess{
				ApplicationID:   app.ApplicationID,
				ApplicationName: app.ApplicationName,
				Permissions:     app.Permissions,
			})
		}

		input := &admin.UpdateServiceKeyInput{
			ID:           serviceKeyID,
			Name:         req.Name,
			Applications: applications,
			ExpiresAt:    req.ExpiresAt,
		}

		output, err := admin.UpdateServiceKey(r.Context(), serviceKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		output.ServiceKey.TokenValue = ""

		httputil.Success(w, http.StatusOK, output.ServiceKey)
	}
}

// DeleteServiceKey deletes a service key
func DeleteServiceKey(serviceKeyRepo admin.ServiceKeyDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceKeyID := chi.URLParam(r, "service_key_id")
		if serviceKeyID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service_key_id is required"))
			return
		}

		input := &admin.DeleteServiceKeyInput{
			ID: serviceKeyID,
		}

		if err := admin.DeleteServiceKey(r.Context(), serviceKeyRepo, input); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusNoContent, nil)
	}
}

// ValidateAccessRequest
type ValidateAccessRequest struct {
	Token           string `json:"token"`
	ApplicationName string `json:"application_name"`
}

func (r ValidateAccessRequest) Validate() error {
	if r.Token == "" {
		return fmt.Errorf("token is required")
	}
	if r.ApplicationName == "" {
		return fmt.Errorf("application_name is required")
	}
	return nil
}

// ValidateAccess validates a service key for a specific application
func ValidateAccess(serviceKeyRepo core.ServiceKeyTokenFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from header
		orgID := r.Header.Get("x-org-id")
		if orgID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("x-org-id header is required"))
			return
		}

		req, err := httputil.Decode[ValidateAccessRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &core.ValidateAccessInput{
			Token:           req.Token,
			ApplicationName: req.ApplicationName,
			OrganizationID:  orgID,
		}

		output, err := core.ValidateAccess(r.Context(), serviceKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, output)
	}
}
