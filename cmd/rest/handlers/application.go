package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/brianfromlife/baluster/internal/core/admin"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/go-chi/chi/v5"
)

// CreateApplicationRequest represents the HTTP request to create an application
type CreateApplicationRequest struct {
	Name        string   `json:"name"` // use underscores instead of spaces (e.g., user_service)
	Description string   `json:"description"`
	Permissions []string `json:"permissions"` // e.g., ["address.read", "access.all", "admin", "basic"]
}

// Validate validates the CreateApplicationRequest
func (r CreateApplicationRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// CreateApplication creates a new application to add seriv
func CreateApplication(appRepo admin.ApplicationCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[CreateApplicationRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &admin.CreateApplicationInput{
			Name:        req.Name,
			Description: req.Description,
			Permissions: req.Permissions,
		}

		_, err = admin.CreateApplication(r.Context(), appRepo, input)
		if err != nil {
			if errors.Is(err, admin.ErrUserInfoNotFound) {
				httputil.Error(w, http.StatusUnauthorized, err)
				return
			}
			if errors.Is(err, admin.ErrApplicationLimitExceeded) {
				httputil.Error(w, http.StatusForbidden, err)
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusCreated, nil)
	}
}

// ListApplications lists applications for an organization
func ListApplications(appRepo admin.ApplicationLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := &admin.ListApplicationsInput{}

		output, err := admin.ListApplications(r.Context(), appRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, map[string]any{"applications": output.Applications})
	}
}

// GetApplication gets an application by ID
func GetApplication(appRepo admin.ApplicationGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		applicationID := chi.URLParam(r, "application_id")
		if applicationID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("application_id is required"))
			return
		}

		input := &admin.GetApplicationInput{
			ID: applicationID,
		}

		output, err := admin.GetApplication(r.Context(), appRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, output.Application)
	}
}

// GetApplicationHistory gets audit history for an application by ID
func GetApplicationHistory(appRepo admin.ApplicationGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		applicationID := chi.URLParam(r, "application_id")
		if applicationID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("application_id is required"))
			return
		}

		input := &admin.GetApplicationHistoryInput{
			ID: applicationID,
		}

		output, err := admin.GetApplicationHistory(r.Context(), appRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, map[string]any{
			"history": output.History,
		})
	}
}

// UpdateApplicationRequest represents the HTTP request to update an application
type UpdateApplicationRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Validate validates the UpdateApplicationRequest
func (r UpdateApplicationRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// UpdateApplication updates an application
func UpdateApplication(appRepo admin.ApplicationUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		applicationID := chi.URLParam(r, "application_id")
		if applicationID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("application_id is required"))
			return
		}

		req, err := httputil.Decode[UpdateApplicationRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &admin.UpdateApplicationInput{
			ID:          applicationID,
			Name:        req.Name,
			Description: req.Description,
			Permissions: req.Permissions,
		}

		output, err := admin.UpdateApplication(r.Context(), appRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, output.Application)
	}
}
