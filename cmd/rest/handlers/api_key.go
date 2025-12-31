package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brianfromlife/baluster/internal/core/admin"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/go-chi/chi/v5"
)

// CreateApiKeyRequest represents the HTTP request to create an API key
type CreateApiKeyRequest struct {
	ApplicationID string     `json:"application_id"`
	Name          string     `json:"name"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

// Validate validates the CreateApiKeyRequest
func (r CreateApiKeyRequest) Validate() error {
	if r.ApplicationID == "" {
		return fmt.Errorf("application_id is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// CreateApiKeyResponse represents the HTTP response with the token value
type CreateApiKeyResponse struct {
	Token      any    `json:"token"` // *types.ApiKey
	TokenValue string `json:"token_value"`
}

// CreateApiKey creates a new API key
func CreateApiKey(apiKeyRepo admin.ApiKeyCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.Decode[CreateApiKeyRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &admin.CreateApiKeyInput{
			ApplicationID: req.ApplicationID,
			Name:          req.Name,
			ExpiresAt:     req.ExpiresAt,
		}

		output, err := admin.CreateApiKey(r.Context(), apiKeyRepo, input)
		if err != nil {
			if errors.Is(err, admin.ErrUserInfoNotFound) {
				httputil.Error(w, http.StatusUnauthorized, err)
				return
			}
			if errors.Is(err, admin.ErrApiKeyLimitExceeded) {
				httputil.Error(w, http.StatusForbidden, err)
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		responseToken := output.Token
		responseToken.TokenValue = ""

		httputil.Success(w, http.StatusCreated, CreateApiKeyResponse{
			Token:      responseToken,
			TokenValue: output.TokenValue,
		})
	}
}

// GetApiKey gets an API key by ID
func GetApiKey(apiKeyRepo admin.ApiKeyGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := chi.URLParam(r, "token_id")
		if tokenID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("token_id is required"))
			return
		}

		input := &admin.GetApiKeyInput{
			ID: tokenID,
		}

		output, err := admin.GetApiKey(r.Context(), apiKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		output.Token.TokenValue = ""

		httputil.Success(w, http.StatusOK, output.Token)
	}
}

// GetApiKeyHistory gets audit history for an API key by ID
func GetApiKeyHistory(apiKeyRepo admin.ApiKeyGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := chi.URLParam(r, "token_id")
		if tokenID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("token_id is required"))
			return
		}

		input := &admin.GetApiKeyHistoryInput{
			ID: tokenID,
		}

		output, err := admin.GetApiKeyHistory(r.Context(), apiKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusOK, map[string]any{
			"history": output.History,
		})
	}
}

// ListApiKeys lists API keys for an organization
func ListApiKeys(apiKeyRepo admin.ApiKeyLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := &admin.ListApiKeysInput{}

		output, err := admin.ListApiKeys(r.Context(), apiKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear token values from response
		for _, token := range output.Tokens {
			token.TokenValue = ""
		}

		httputil.Success(w, http.StatusOK, output.Tokens)
	}
}

// UpdateApiKeyRequest represents the HTTP request to update an API key
type UpdateApiKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Validate validates the UpdateApiKeyRequest
func (r UpdateApiKeyRequest) Validate() error {
	return nil
}

// UpdateApiKey updates an API key
func UpdateApiKey(apiKeyRepo admin.ApiKeyUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := chi.URLParam(r, "token_id")
		if tokenID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("token_id is required"))
			return
		}

		req, err := httputil.Decode[UpdateApiKeyRequest](r)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}

		input := &admin.UpdateApiKeyInput{
			ID:        tokenID,
			Name:      req.Name,
			ExpiresAt: req.ExpiresAt,
		}

		output, err := admin.UpdateApiKey(r.Context(), apiKeyRepo, input)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		// Clear the hashed value from response
		output.Token.TokenValue = ""

		httputil.Success(w, http.StatusOK, output.Token)
	}
}

// DeleteApiKey deletes an API key
func DeleteApiKey(apiKeyRepo admin.ApiKeyDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := chi.URLParam(r, "token_id")
		if tokenID == "" {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("token_id is required"))
			return
		}

		input := &admin.DeleteApiKeyInput{
			ID: tokenID,
		}

		if err := admin.DeleteApiKey(r.Context(), apiKeyRepo, input); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		httputil.Success(w, http.StatusNoContent, nil)
	}
}
