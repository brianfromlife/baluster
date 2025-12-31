package auth

import (
	"context"

	"github.com/brianfromlife/baluster/internal/storage"
)

type ApiKeyValidator struct {
	apiKeyRepo *storage.ApiKeyRepository
}

func NewApiKeyValidator(apiKeyRepo *storage.ApiKeyRepository) *ApiKeyValidator {
	return &ApiKeyValidator{
		apiKeyRepo: apiKeyRepo,
	}
}

// Validate validates an API key and returns whether it's valid
func (v *ApiKeyValidator) Validate(ctx context.Context, tokenValue string) (bool, error) {
	token, err := v.apiKeyRepo.FindByTokenValue(ctx, tokenValue)
	if err != nil {
		return false, nil
	}

	if token.IsExpired() {
		return false, nil
	}

	return true, nil
}
