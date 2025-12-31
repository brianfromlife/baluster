package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, status int, err error) {
	resp := ErrorResponse{
		Error:   http.StatusText(status),
		Message: err.Error(),
	}
	JSON(w, status, resp)
}

// Success writes a success JSON response
func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, data)
}
