package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/core"
	"github.com/brianfromlife/baluster/internal/core/admin"
	"github.com/brianfromlife/baluster/internal/types"
	"github.com/go-chi/chi/v5"
)

// Ensure mocks implement the interfaces
var (
	_ admin.OrganizationCreator  = (*mockOrganizationRepo)(nil)
	_ admin.ApplicationCreator   = (*mockApplicationRepo)(nil)
	_ admin.ApplicationLister    = (*mockApplicationRepo)(nil)
	_ admin.ApplicationGetter    = (*mockApplicationRepo)(nil)
	_ admin.ApplicationUpdater   = (*mockApplicationRepo)(nil)
	_ admin.ApiKeyCreator        = (*mockApiKeyRepo)(nil)
	_ admin.ApiKeyGetter         = (*mockApiKeyRepo)(nil)
	_ admin.ApiKeyLister         = (*mockApiKeyRepo)(nil)
	_ admin.ApiKeyUpdater        = (*mockApiKeyRepo)(nil)
	_ admin.ApiKeyDeleter        = (*mockApiKeyRepo)(nil)
	_ admin.ServiceKeyCreator    = (*mockServiceKeyRepo)(nil)
	_ admin.ServiceKeyGetter     = (*mockServiceKeyRepo)(nil)
	_ admin.ServiceKeyLister     = (*mockServiceKeyRepo)(nil)
	_ admin.ServiceKeyUpdater    = (*mockServiceKeyRepo)(nil)
	_ admin.ServiceKeyDeleter    = (*mockServiceKeyRepo)(nil)
	_ core.ServiceKeyTokenFinder = (*mockServiceKeyRepo)(nil)
)

// Mock Organization Repository

type mockOrganizationRepo struct {
	organizations []*types.Organization
	createErr     error
}

func (m *mockOrganizationRepo) Create(ctx context.Context, org *types.Organization) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.organizations = append(m.organizations, org)
	return nil
}

// Mock Organization Member Repository

type mockOrganizationMemberRepo struct {
	addMemberErr error
}

func (m *mockOrganizationMemberRepo) AddMember(ctx context.Context, orgID, userID string) error {
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	return nil
}

func (m *mockOrganizationMemberRepo) CreateOrganizationWithMember(ctx context.Context, org *types.Organization, userID string) error {
	return nil
}

// Mock Application Repository

type mockApplicationRepo struct {
	applications []*types.Application
	createErr    error
	listErr      error
	countErr     error
}

func (m *mockApplicationRepo) Create(ctx context.Context, app *types.Application, userID, githubID, username string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.applications = append(m.applications, app)
	return nil
}

func (m *mockApplicationRepo) CountByOrganization(ctx context.Context, organizationID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	count := 0
	for _, app := range m.applications {
		if app.OrganizationID == organizationID {
			count++
		}
	}
	return count, nil
}

func (m *mockApplicationRepo) ListByOrganization(ctx context.Context, organizationID string) ([]*types.Application, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*types.Application
	for _, app := range m.applications {
		if app.OrganizationID == organizationID {
			result = append(result, app)
		}
	}
	return result, nil
}

func (m *mockApplicationRepo) Get(ctx context.Context, organizationID, id string) (*types.Application, error) {
	for _, app := range m.applications {
		if app.ID == id && app.OrganizationID == organizationID {
			return app, nil
		}
	}
	return nil, nil
}

func (m *mockApplicationRepo) GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error) {
	return []*types.AuditHistory{}, nil
}

func (m *mockApplicationRepo) Update(ctx context.Context, app *types.Application, userID, githubID, username string) error {
	for i, existingApp := range m.applications {
		if existingApp.ID == app.ID {
			m.applications[i] = app
			return nil
		}
	}
	return nil
}

// Mock API Key Repository

type mockApiKeyRepo struct {
	apiKeys   []*types.ApiKey
	createErr error
	getErr    error
	listErr   error
	countErr  error
	updateErr error
	deleteErr error
}

func (m *mockApiKeyRepo) Create(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.apiKeys = append(m.apiKeys, apiKey)
	return nil
}

func (m *mockApiKeyRepo) CountByOrganization(ctx context.Context, organizationID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	count := 0
	for _, key := range m.apiKeys {
		if key.OrganizationID == organizationID {
			count++
		}
	}
	return count, nil
}

func (m *mockApiKeyRepo) Get(ctx context.Context, organizationID, id string) (*types.ApiKey, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, key := range m.apiKeys {
		if key.ID == id && key.OrganizationID == organizationID {
			return key, nil
		}
	}
	return nil, nil
}

func (m *mockApiKeyRepo) GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error) {
	return []*types.AuditHistory{}, nil
}

func (m *mockApiKeyRepo) ListByOrganization(ctx context.Context, organizationID string) ([]*types.ApiKey, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*types.ApiKey
	for _, key := range m.apiKeys {
		if key.OrganizationID == organizationID {
			result = append(result, key)
		}
	}
	return result, nil
}

func (m *mockApiKeyRepo) Update(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, key := range m.apiKeys {
		if key.ID == apiKey.ID {
			m.apiKeys[i] = apiKey
			return nil
		}
	}
	return nil
}

func (m *mockApiKeyRepo) Delete(ctx context.Context, apiKey *types.ApiKey, userID, githubID, username string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, key := range m.apiKeys {
		if key.ID == apiKey.ID {
			m.apiKeys = append(m.apiKeys[:i], m.apiKeys[i+1:]...)
			return nil
		}
	}
	return nil
}

// Mock Service Key Repository

type mockServiceKeyRepo struct {
	serviceKeys []*types.ServiceKey
	createErr   error
	getErr      error
	listErr     error
	countErr    error
	updateErr   error
	deleteErr   error
	findErr     error
}

func (m *mockServiceKeyRepo) Create(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.serviceKeys = append(m.serviceKeys, serviceKey)
	return nil
}

func (m *mockServiceKeyRepo) CountByOrganization(ctx context.Context, organizationID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	count := 0
	for _, key := range m.serviceKeys {
		if key.OrganizationID == organizationID {
			count++
		}
	}
	return count, nil
}

func (m *mockServiceKeyRepo) Get(ctx context.Context, organizationID, id string) (*types.ServiceKey, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, key := range m.serviceKeys {
		if key.ID == id && key.OrganizationID == organizationID {
			return key, nil
		}
	}
	return nil, nil
}

func (m *mockServiceKeyRepo) GetHistory(ctx context.Context, organizationID, entityID string) ([]*types.AuditHistory, error) {
	return []*types.AuditHistory{}, nil
}

func (m *mockServiceKeyRepo) ListByOrganization(ctx context.Context, organizationID string) ([]*types.ServiceKey, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*types.ServiceKey
	for _, key := range m.serviceKeys {
		if key.OrganizationID == organizationID {
			result = append(result, key)
		}
	}
	return result, nil
}

func (m *mockServiceKeyRepo) Update(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, key := range m.serviceKeys {
		if key.ID == serviceKey.ID {
			m.serviceKeys[i] = serviceKey
			return nil
		}
	}
	return nil
}

func (m *mockServiceKeyRepo) Delete(ctx context.Context, serviceKey *types.ServiceKey, userID, githubID, username string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, key := range m.serviceKeys {
		if key.ID == serviceKey.ID {
			m.serviceKeys = append(m.serviceKeys[:i], m.serviceKeys[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockServiceKeyRepo) FindByTokenValue(ctx context.Context, tokenValue string) (*types.ServiceKey, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, key := range m.serviceKeys {
		if key.TokenValue == tokenValue {
			return key, nil
		}
	}
	return nil, nil
}

func (m *mockServiceKeyRepo) FindByTokenValueInOrg(ctx context.Context, organizationID, tokenValue string) (*types.ServiceKey, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, key := range m.serviceKeys {
		if key.TokenValue == tokenValue && key.OrganizationID == organizationID {
			return key, nil
		}
	}
	return nil, nil
}

// Test helpers

func newTestRequest(method, path string, body any) *http.Request {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}
	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func withUserContext(req *http.Request, userID, githubID, username string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, auth.UserIDKey, userID)
	ctx = context.WithValue(ctx, auth.GitHubIDKey, githubID)
	ctx = context.WithValue(ctx, auth.UsernameKey, username)
	return req.WithContext(ctx)
}

func withOrgContext(req *http.Request, orgID string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, auth.OrganizationIDKey, orgID)
	req.Header.Set("x-org-id", orgID)
	return req.WithContext(ctx)
}
