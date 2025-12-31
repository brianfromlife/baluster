package admin

import "errors"

var (
	// ErrUserInfoNotFound is returned when user information is not found in context
	ErrUserInfoNotFound = errors.New("user information not found in context")
	// ErrApplicationLimitExceeded is returned when the maximum number of applications (20) is reached
	ErrApplicationLimitExceeded = errors.New("maximum number of applications (20) reached")
	// ErrServiceKeyLimitExceeded is returned when the maximum number of service keys (50) is reached
	ErrServiceKeyLimitExceeded = errors.New("maximum number of service keys (50) reached")
	// ErrApiKeyLimitExceeded is returned when the maximum number of API keys (50) is reached
	ErrApiKeyLimitExceeded = errors.New("maximum number of API keys (50) reached")
)
