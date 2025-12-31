package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianfromlife/baluster/internal/types"
)

func TestCreateServiceKey(t *testing.T) {
	tests := []struct {
		name           string
		body           CreateServiceKeyRequest
		repo           *mockServiceKeyRepo
		expectedStatus int
	}{
		{
			name: "valid request",
			body: CreateServiceKeyRequest{
				Name: "Test Service Key",
				Applications: []ApplicationAccessRequest{
					{ApplicationID: "app-1", ApplicationName: "test_app", Permissions: []string{"read"}},
				},
			},
			repo:           &mockServiceKeyRepo{},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			body: CreateServiceKeyRequest{
				Name: "",
			},
			repo:           &mockServiceKeyRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too short",
			body: CreateServiceKeyRequest{
				Name: "ab",
				Applications: []ApplicationAccessRequest{
					{ApplicationID: "app-1", ApplicationName: "test_app", Permissions: []string{"read"}},
				},
			},
			repo:           &mockServiceKeyRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service key limit exceeded",
			body: CreateServiceKeyRequest{
				Name: "Test Service Key",
				Applications: []ApplicationAccessRequest{
					{ApplicationID: "app-1", ApplicationName: "test_app", Permissions: []string{"read"}},
				},
			},
			repo: &mockServiceKeyRepo{
				serviceKeys: make([]*types.ServiceKey, 50), // Already at limit
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize service keys with org-1 if needed
			if tt.repo != nil && len(tt.repo.serviceKeys) > 0 {
				for i := range tt.repo.serviceKeys {
					tt.repo.serviceKeys[i] = &types.ServiceKey{
						ID:             "sk-existing",
						OrganizationID: "org-1",
					}
				}
			}
			handler := CreateServiceKey(tt.repo)

			req := newTestRequest(http.MethodPost, "/service-keys", tt.body)
			// Add user context for valid requests (they need auth)
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

func TestGetServiceKey(t *testing.T) {
	repo := &mockServiceKeyRepo{
		serviceKeys: []*types.ServiceKey{
			{ID: "sk-1", OrganizationID: "org-1", Name: "Test Service Key"},
		},
	}

	handler := GetServiceKey(repo)
	req := newTestRequest(http.MethodGet, "/service-keys/sk-1?organization_id=org-1", nil)
	req = withURLParam(req, "service_key_id", "sk-1")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestListServiceKeys(t *testing.T) {
	repo := &mockServiceKeyRepo{
		serviceKeys: []*types.ServiceKey{
			{ID: "1", OrganizationID: "org-1", Name: "Key 1"},
			{ID: "2", OrganizationID: "org-1", Name: "Key 2"},
			{ID: "3", OrganizationID: "org-2", Name: "Key 3"},
		},
	}

	handler := ListServiceKeys(repo)
	req := newTestRequest(http.MethodGet, "/organizations/org-1/service-keys", nil)
	req = withURLParam(req, "organization_id", "org-1")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var keys []*types.ServiceKey
	if err := json.NewDecoder(rr.Body).Decode(&keys); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 service keys, got %d", len(keys))
	}
}

func TestUpdateServiceKey(t *testing.T) {
	repo := &mockServiceKeyRepo{
		serviceKeys: []*types.ServiceKey{
			{ID: "sk-1", OrganizationID: "org-1", Name: "Old Name"},
		},
	}

	handler := UpdateServiceKey(repo)
	req := newTestRequest(http.MethodPut, "/service-keys/sk-1?organization_id=org-1", UpdateServiceKeyRequest{
		Name: "New Name",
	})
	req = withURLParam(req, "service_key_id", "sk-1")
	req = withUserContext(req, "user-1", "github-123", "testuser")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDeleteServiceKey(t *testing.T) {
	repo := &mockServiceKeyRepo{
		serviceKeys: []*types.ServiceKey{
			{ID: "sk-1", OrganizationID: "org-1", Name: "Test Service Key"},
		},
	}

	handler := DeleteServiceKey(repo)
	req := newTestRequest(http.MethodDelete, "/service-keys/sk-1?organization_id=org-1", nil)
	req = withURLParam(req, "service_key_id", "sk-1")
	req = withUserContext(req, "user-1", "github-123", "testuser")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestValidateAccess(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	repo := &mockServiceKeyRepo{
		serviceKeys: []*types.ServiceKey{
			{
				ID:             "sk-1",
				OrganizationID: "org-1",
				TokenValue:     "valid-token",
				ExpiresAt:      &expiry,
				Applications: []types.ApplicationAccess{
					{ApplicationID: "app-1", ApplicationName: "test_app", Permissions: []string{"read", "write"}},
				},
			},
		},
	}

	tests := []struct {
		name           string
		body           ValidateAccessRequest
		expectedStatus int
	}{
		{
			name: "valid token and application",
			body: ValidateAccessRequest{
				Token:           "valid-token",
				ApplicationName: "test_app",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing token",
			body: ValidateAccessRequest{
				ApplicationName: "test_app",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing application_name",
			body: ValidateAccessRequest{
				Token: "valid-token",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ValidateAccess(repo)

			req := newTestRequest(http.MethodPost, "/api/validate/access", tt.body)
			// ValidateAccess requires x-org-id header
			if tt.expectedStatus == http.StatusOK {
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
