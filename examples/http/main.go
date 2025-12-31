// Example client for validating access with the Baluster API
//
// This program demonstrates how to use the Baluster API to validate service key access
// for a specific application. It sends a POST request to the /api/v1/access endpoint
// with a service key token and application name.
//
// IMPORTANT: The x-org-id header is REQUIRED when calling the /api/v1/access endpoint.
// You can find your organization ID in the response when creating a service key,
// or by checking your organization details in the Baluster dashboard.
//
// Building:
//
//	go build -o bin/http_example ./examples/http
//
// Running:
//
//	./bin/http_example
//
// Configuration:
//
//	Edit the constants at the top of this file to set your:
//	- accessKey: Your API key (Bearer token) for authenticating with Baluster
//	- serviceKey: The service key token to validate
//	- applicationName: The name of the application to check access for
//	- organizationID: Your organization ID (required in x-org-id header)
//	- baseURL: Base URL of the API server (default: http://localhost:8080)
//
// HTTP Client Instructions:
//
// When making requests to /api/v1/access, include the following headers:
//
//	Authorization: Bearer <your-api-key>
//	x-org-id: <your-organization-id>
//	Content-Type: application/json

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Hardcode your keys here
const (
	accessKey       = "syOSmb7RR1hpT7CYRt-SoD8FGbIfee6cKNXkNgaWHUY="
	serviceKey      = "JKLwf3U0H9yzhOuExupLBedp6pYc1gH6LmG0w3LDGMI="
	applicationName = "mail_service"
	organizationID  = "yudW2rZOmZaV8lX7cfafpA=="
	baseURL         = "http://localhost:8080"
)

type ValidateAccessRequest struct {
	Token           string `json:"token"`
	ApplicationName string `json:"application_name"`
}

type ValidateAccessResponse struct {
	Valid          bool     `json:"valid"`
	ServiceKeyID   string   `json:"service_key_id"`
	OrganizationID string   `json:"organization_id"`
	ApplicationID  string   `json:"application_id"`
	Permissions    []string `json:"permissions"`
	ExpiresAt      int64    `json:"expires_at"`
}

func main() {
	// Prepare the request body
	reqBody := ValidateAccessRequest{
		Token:           serviceKey,
		ApplicationName: applicationName,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request body: %v\n", err)
		os.Exit(1)
	}

	// Create the request
	url := fmt.Sprintf("%s/api/v1/access", baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessKey))
	req.Header.Set("x-org-id", organizationID) // REQUIRED: Organization ID header

	// Make the request
	fmt.Printf("Making request to: %s\n", url)
	fmt.Printf("Authorization: Bearer %s\n", maskKey(accessKey))
	fmt.Printf("x-org-id: %s\n", organizationID)
	fmt.Printf("Request body: %s\n", string(jsonBody))
	fmt.Println()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Display response
	fmt.Printf("Response Status: %s (%d)\n", resp.Status, resp.StatusCode)
	fmt.Println()

	if resp.StatusCode == http.StatusOK {
		var accessResp ValidateAccessResponse
		if err := json.Unmarshal(body, &accessResp); err != nil {
			fmt.Printf("Response body (raw): %s\n", string(body))
			fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Access validation successful!")
		fmt.Printf("Valid: %v\n", accessResp.Valid)
		fmt.Printf("Permissions: %v\n", accessResp.Permissions)
		if accessResp.ExpiresAt > 0 {
			fmt.Printf("Expires At: %d (Unix timestamp)\n", accessResp.ExpiresAt)
		} else {
			fmt.Println("Expires At: Never")
		}
	} else {
		fmt.Printf("❌ Request failed\n")
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}
}

// maskKey masks most of the key for display purposes
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
