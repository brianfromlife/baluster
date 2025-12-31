package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/brianfromlife/baluster/internal/core/admin"
	httputil "github.com/brianfromlife/baluster/internal/http"
)

// CreateOrganizationRequest represents the HTTP request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name"`
}

// Validate validates the CreateOrganizationRequest
func (r CreateOrganizationRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	return nil
}

// CreateOrganization creates a new organization
func CreateOrganization(orgRepo admin.OrganizationCreator, memberRepo admin.OrganizationMemberAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[CreateOrganizationRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &admin.CreateOrganizationInput{
			Name: req.Name,
		}

		output, err := admin.CreateOrganization(r.Context(), orgRepo, memberRepo, input)
		if err != nil {
			if errors.Is(err, admin.ErrUserInfoNotFound) {
				httputil.Error(w, http.StatusUnauthorized, err)
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusCreated, output.Organization)
	}
}
