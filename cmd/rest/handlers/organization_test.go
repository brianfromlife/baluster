package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateOrganization(t *testing.T) {
	tests := []struct {
		name           string
		body           CreateOrganizationRequest
		expectedStatus int
	}{
		{
			name: "valid request",
			body: CreateOrganizationRequest{
				Name: "Test Org",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			body: CreateOrganizationRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too short",
			body: CreateOrganizationRequest{
				Name: "ab",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgRepo := &mockOrganizationRepo{}
			memberRepo := &mockOrganizationMemberRepo{}
			handler := CreateOrganization(orgRepo, memberRepo)

			req := newTestRequest(http.MethodPost, "/organizations", tt.body)
			// Add user context for valid requests (they need auth)
			if tt.expectedStatus == http.StatusCreated {
				req = withUserContext(req, "user-1", "github-123", "testuser")
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
