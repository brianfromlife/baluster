package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianfromlife/baluster/internal/types"
)

func TestCreateApiKey(t *testing.T) {
	tests := []struct {
		name           string
		body           CreateApiKeyRequest
		repo           *mockApiKeyRepo
		expectedStatus int
	}{
		{
			name: "valid request",
			body: CreateApiKeyRequest{
				ApplicationID: "app-1",
				Name:          "Test API Key",
			},
			repo:           &mockApiKeyRepo{},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "valid request with expiration",
			body: CreateApiKeyRequest{
				ApplicationID: "app-1",
				Name:          "Test API Key",
				ExpiresAt:     func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			},
			repo:           &mockApiKeyRepo{},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			body: CreateApiKeyRequest{
				ApplicationID: "app-1",
			},
			repo:           &mockApiKeyRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing application_id",
			body: CreateApiKeyRequest{
				Name: "Test API Key",
			},
			repo:           &mockApiKeyRepo{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "API key limit exceeded",
			body: CreateApiKeyRequest{
				ApplicationID: "app-1",
				Name:          "Test API Key",
			},
			repo: &mockApiKeyRepo{
				apiKeys: make([]*types.ApiKey, 50), // Already at limit
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize API keys with org-1 if needed
			if tt.repo != nil && len(tt.repo.apiKeys) > 0 {
				for i := range tt.repo.apiKeys {
					tt.repo.apiKeys[i] = &types.ApiKey{
						ID:             "key-existing",
						OrganizationID: "org-1",
					}
				}
			}
			handler := CreateApiKey(tt.repo)

			req := newTestRequest(http.MethodPost, "/api-keys", tt.body)
			// Add user context and organization context for valid requests (they need auth and org membership)
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

func TestGetApiKey(t *testing.T) {
	repo := &mockApiKeyRepo{
		apiKeys: []*types.ApiKey{
			{ID: "key-1", OrganizationID: "org-1", Name: "Test Key"},
		},
	}

	handler := GetApiKey(repo)
	req := newTestRequest(http.MethodGet, "/api-keys/key-1", nil)
	req = withURLParam(req, "token_id", "key-1")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestListApiKeys(t *testing.T) {
	repo := &mockApiKeyRepo{
		apiKeys: []*types.ApiKey{
			{ID: "1", OrganizationID: "org-1", Name: "Key 1"},
			{ID: "2", OrganizationID: "org-1", Name: "Key 2"},
			{ID: "3", OrganizationID: "org-2", Name: "Key 3"},
		},
	}

	handler := ListApiKeys(repo)
	req := newTestRequest(http.MethodGet, "/api-keys", nil)
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var keys []*types.ApiKey
	if err := json.NewDecoder(rr.Body).Decode(&keys); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 API keys, got %d", len(keys))
	}
}

func TestUpdateApiKey(t *testing.T) {
	repo := &mockApiKeyRepo{
		apiKeys: []*types.ApiKey{
			{ID: "key-1", OrganizationID: "org-1", Name: "Old Name"},
		},
	}

	handler := UpdateApiKey(repo)
	req := newTestRequest(http.MethodPut, "/api-keys/key-1", UpdateApiKeyRequest{
		Name: "New Name",
	})
	req = withURLParam(req, "token_id", "key-1")
	req = withUserContext(req, "user-1", "github-123", "testuser")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDeleteApiKey(t *testing.T) {
	repo := &mockApiKeyRepo{
		apiKeys: []*types.ApiKey{
			{ID: "key-1", OrganizationID: "org-1", Name: "Test Key"},
		},
	}

	handler := DeleteApiKey(repo)
	req := newTestRequest(http.MethodDelete, "/api-keys/key-1", nil)
	req = withURLParam(req, "token_id", "key-1")
	req = withUserContext(req, "user-1", "github-123", "testuser")
	req = withOrgContext(req, "org-1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}
