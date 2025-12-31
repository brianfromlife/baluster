// Example gRPC client for validating access with the Baluster API
//
// This program demonstrates how to use the Baluster gRPC API to validate service key access
// for a specific application. It sends a ValidateAccess RPC call to the AccessService
// with a service key token and application name.
//
// IMPORTANT: The organization_id is REQUIRED when calling ValidateAccess.
// You can find your organization ID in the response when creating a service key,
// or by checking your organization details in the Baluster dashboard.
//
// Building:
//
//	go build -o bin/grpc_example ./examples/grpc
//
// Running:
//
//	./bin/grpc_example
//
// Configuration:
//
//	Edit the constants at the top of this file to set your:
//	- accessKey: Your API key (Bearer token) for authenticating with Baluster
//	- serviceKey: The service key token to validate
//	- applicationName: The name of the application to check access for
//	- organizationID: Your organization ID (required in the request)
//	- baseURL: Base URL of the gRPC server (default: http://localhost:5050)
//
// gRPC Client Instructions:
//
// When making requests to ValidateAccess, include the following:
//
//	Authorization: Bearer <your-api-key> (in request headers)
//	organization_id: <your-organization-id> (in the request message)
//
// Example with grpcurl:
//
//	grpcurl -plaintext \
//	  -H "Authorization: Bearer <your-api-key>" \
//	  -d '{"token": "<service-key>", "application_name": "my-app", "organization_id": "<org-id>"}' \
//	  localhost:5050 baluster.v1.AccessService/ValidateAccess

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"

	balusterv1 "github.com/brianfromlife/baluster/internal/gen"
	balusterv1connect "github.com/brianfromlife/baluster/internal/gen/balusterv1connect"
)

// Hardcode your keys here
const (
	accessKey       = "syOSmb7RR1hpT7CYRt-SoD8FGbIfee6cKNXkNgaWHUY="
	serviceKey      = "JKLwf3U0H9yzhOuExupLBedp6pYc1gH6LmG0w3LDGMI="
	applicationName = "mail_service"
	organizationID  = "yudW2rZOmZaV8lX7cfafpA=="
	baseURL         = "http://localhost:5050"
)

func main() {
	// Create HTTP client with authorization header
	httpClient := &http.Client{}
	client := balusterv1connect.NewAccessServiceClient(
		httpClient,
		baseURL,
		connect.WithInterceptors(newAuthInterceptor(accessKey)),
	)

	// Create the request
	req := connect.NewRequest(&balusterv1.ValidateAccessRequest{
		Token:           serviceKey,
		ApplicationName: applicationName,
		OrganizationId:  organizationID,
	})

	// Make the request
	fmt.Printf("Making gRPC request to: %s\n", baseURL)
	fmt.Printf("Authorization: Bearer %s\n", maskKey(accessKey))
	fmt.Printf("Token: %s\n", maskKey(serviceKey))
	fmt.Printf("Application Name: %s\n", applicationName)
	fmt.Printf("Organization ID: %s\n", organizationID)
	fmt.Println()

	ctx := context.Background()
	resp, err := client.ValidateAccess(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request failed: %v\n", err)
		os.Exit(1)
	}

	// Display response
	fmt.Println("✅ Access validation successful!")
	fmt.Printf("Valid: %v\n", resp.Msg.Valid)
	fmt.Printf("Permissions: %v\n", resp.Msg.Permissions)
}

// authInterceptor adds the Authorization header to requests
type authInterceptor struct {
	accessKey string
}

func newAuthInterceptor(accessKey string) connect.UnaryInterceptorFunc {
	interceptor := &authInterceptor{accessKey: accessKey}
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", fmt.Sprintf("Bearer %s", interceptor.accessKey))
			return next(ctx, req)
		}
	}
}

// maskKey masks most of the key for display purposes
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
