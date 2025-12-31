package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/brianfromlife/baluster/internal/types"
)

type ContextKey string

const (
	UserIDKey         ContextKey = "user_id"
	GitHubIDKey       ContextKey = "github_id"
	UsernameKey       ContextKey = "username"
	OrganizationIDKey ContextKey = "organization_id"
)

// TokenRefreshHeader is the header name for the refreshed token
const TokenRefreshHeader = "X-Refreshed-Token"

// UserGetter is an interface for retrieving user information needed for token refresh
type UserGetter interface {
	GetByID(ctx context.Context, id string, githubID string) (*types.User, error)
}

// OrganizationMemberChecker is an interface for checking organization membership
type OrganizationMemberChecker interface {
	IsMember(ctx context.Context, orgID, userID string) (bool, error)
}

// JWTAuthMiddleware creates middleware for JWT authentication with token refresh support
// If userRepo is provided, expired tokens will be automatically refreshed
func JWTAuthMiddleware(cfg JWTConfig, userRepo UserGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(cfg, parts[1])
			if err != nil {
				// Check if token is expired and we can refresh it
				if errors.Is(err, ErrTokenExpired) && claims != nil {
					logger := slog.Default().With(
						"user_id", claims.UserID,
						"github_id", claims.GitHubID,
					)

					logger.Info("token expired, attempting refresh")

					// Look up user to get current information
					user, lookupErr := userRepo.GetByID(r.Context(), claims.UserID, claims.GitHubID)
					if lookupErr != nil {
						logger.Error("failed to lookup user for token refresh", "error", lookupErr)
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}

					// Generate new token
					newToken, genErr := GenerateToken(cfg, user.ID, user.GitHubID, user.Username)
					if genErr != nil {
						logger.Error("failed to generate refreshed token", "error", genErr)
						http.Error(w, "failed to refresh token", http.StatusInternalServerError)
						return
					}

					logger.Info("token refreshed successfully")

					// Set refreshed token in response header
					w.Header().Set(TokenRefreshHeader, newToken)

					// Use user information from database for context
					ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
					ctx = context.WithValue(ctx, GitHubIDKey, user.GitHubID)
					ctx = context.WithValue(ctx, UsernameKey, user.Username)

					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				// Token is invalid or expired without refresh capability
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Token is valid
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, GitHubIDKey, claims.GitHubID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			fmt.Println("userId", claims.UserID, "github", claims.GitHubID, "username", claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ApiKeyAuthInterceptor creates a Connect interceptor that validates API keys
// from the Authorization header. This token is used to authenticate gRPC requests to Baluster APIs.
func ApiKeyAuthInterceptor(validator *ApiKeyValidator) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					nil,
				)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					nil,
				)
			}

			tokenValue := parts[1]
			valid, err := validator.Validate(ctx, tokenValue)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeInternal,
					err,
				)
			}

			if !valid {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					nil,
				)
			}

			return next(ctx, req)
		}
	}
}

// ApiKeyAuthMiddleware creates middleware for API key-based authentication
// Validates API keys from the Authorization header used to call Baluster APIs
func ApiKeyAuthMiddleware(validator *ApiKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenValue := parts[1]
			valid, err := validator.Validate(r.Context(), tokenValue)
			if err != nil {
				http.Error(w, "failed to validate token", http.StatusInternalServerError)
				return
			}

			if !valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OrganizationMembershipMiddleware creates middleware that validates organization membership
// It reads the x-org-id header, checks membership with caching, and adds org ID to context
func OrganizationMembershipMiddleware(memberRepo OrganizationMemberChecker, cache *MembershipCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			orgID := r.Header.Get("x-org-id")
			if orgID == "" {
				http.Error(w, "missing x-org-id header", http.StatusUnauthorized)
				return
			}

			// userID is set from the JWT middleware
			userID, ok := GetUserID(r.Context())
			if !ok {
				http.Error(w, "user not authenticated", http.StatusUnauthorized)
				return
			}

			isMember, found := cache.Get(orgID, userID)
			if !found {
				var err error
				isMember, err = memberRepo.IsMember(r.Context(), orgID, userID)
				if err != nil {
					logger := slog.Default().With(
						"org_id", orgID,
						"user_id", userID,
						"error", err,
					)
					logger.Error("failed to check organization membership")
					http.Error(w, "failed to validate organization membership", http.StatusInternalServerError)
					return
				}

				// Cache the result
				cache.Set(orgID, userID, isMember)
			}

			if !isMember {
				http.Error(w, "user is not a member of this organization", http.StatusUnauthorized)
				return
			}

			// Add organization ID to context
			ctx := context.WithValue(r.Context(), OrganizationIDKey, orgID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
