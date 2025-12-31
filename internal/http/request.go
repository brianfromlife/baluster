package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Validatable is an interface for request structs that can be validated
type Validatable interface {
	Validate() error
}

// Decode decodes the request body into the given struct and validates it
func Decode[T Validatable](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		var t T
		return t, fmt.Errorf("invalid request body: %w", err)
	}

	if err := v.Validate(); err != nil {
		var t T
		return t, err
	}

	return v, nil
}
