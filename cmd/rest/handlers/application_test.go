package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianfromlife/baluster/internal/types"
)

func TestCreateApplication(t *testing.T) {
	tests := []struct {
		name           string
		body           CreateApplicationRequest
		repo           *mockApplicationRepo
		expectedStatus int
	}{
		{
			name: "valid request",
			body: CreateApplicationRequest{
				Name:        "test_app",
				Description: "Test Application",
				Permissions: []string{"read", "write"},
			},
			repo:           &mockApplicationRepo{},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			body: CreateApplicationRequest{
				Name: "",
			},
			repo:           &mockApplicationRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "application limit exceeded",
			body: CreateApplicationRequest{
				Name:        "test_app",
				Description: "Test Application",
				Permissions: []string{"read", "write"},
			},
			repo: &mockApplicationRepo{
				applications: make([]*types.Application, 20), // Already at limit
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize applications with org-1 if needed
			if tt.repo != nil && len(tt.repo.applications) > 0 {
				for i := range tt.repo.applications {
					tt.repo.applications[i] = &types.Application{
						ID:             "app-existing",
						OrganizationID: "org-1",
					}
				}
			}
			handler := CreateApplication(tt.repo)

			req := newTestRequest(http.MethodPost, "/applications", tt.body)
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusForbidden {
				req = withUserContext(req, "user-1", "github-123", "testuser")
				req = withOrgContext(req, "org-1")
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestListApplications(t *testing.T) {
	repo := &mockApplicationRepo{
		applications: []*types.Application{
			{ID: "1", OrganizationID: "org-1", Name: "app_1"},
			{ID: "2", OrganizationID: "org-1", Name: "app_2"},
			{ID: "3", OrganizationID: "org-2", Name: "app_3"},
		},
	}

	handler := ListApplications(repo)
	req := newTestRequest(http.MethodGet, "/organizations/org-1/applications", nil)
	req = withURLParam(req, "organization_id", "org-1")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	appsValue, ok := response["applications"]
	if !ok {
		t.Fatalf("applications key not found in response: %v", response)
	}

	// Decode the applications array
	appsBytes, err := json.Marshal(appsValue)
	if err != nil {
		t.Fatalf("failed to marshal applications: %v", err)
	}

	var apps []*types.Application
	if err := json.Unmarshal(appsBytes, &apps); err != nil {
		t.Fatalf("failed to unmarshal applications: %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("expected 2 applications, got %d", len(apps))
	}
}
