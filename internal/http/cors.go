package http

import (
	"net/http"
)

// CORSOptions configures CORS middleware behavior
type CORSOptions struct {
	// AllowedHeaders are additional headers to allow beyond the defaults
	AllowedHeaders []string
	// ExposeHeaders are headers to expose to the client
	ExposeHeaders []string
}

// CORS creates a middleware that handles CORS for all origins
func CORS(opts ...CORSOptions) func(http.Handler) http.Handler {
	var options CORSOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Default allowed headers
	allowedHeaders := []string{"Accept", "Authorization", "Content-Type"}
	allowedHeaders = append(allowedHeaders, options.AllowedHeaders...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					
					// Set allowed headers
					headersStr := ""
					for i, h := range allowedHeaders {
						if i > 0 {
							headersStr += ", "
						}
						headersStr += h
					}
					w.Header().Set("Access-Control-Allow-Headers", headersStr)
					w.Header().Set("Access-Control-Max-Age", "300")
				}
				w.WriteHeader(http.StatusOK)
				return
			}

			// Set CORS headers for actual requests
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				
				// Set expose headers if provided
				if len(options.ExposeHeaders) > 0 {
					exposeStr := ""
					for i, h := range options.ExposeHeaders {
						if i > 0 {
							exposeStr += ", "
						}
						exposeStr += h
					}
					w.Header().Set("Access-Control-Expose-Headers", exposeStr)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

