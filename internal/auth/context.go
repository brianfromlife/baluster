package auth

import (
	"context"
)

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetGitHubID retrieves the GitHub ID from context
func GetGitHubID(ctx context.Context) (string, bool) {
	githubID, ok := ctx.Value(GitHubIDKey).(string)
	return githubID, ok
}

// GetUsername retrieves the username from context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}

// GetUserInfo retrieves all user information from context
// Returns userID, githubID, username, and a boolean indicating if all values were found
func GetUserInfo(ctx context.Context) (userID, githubID, username string, ok bool) {
	userID, userIDOk := GetUserID(ctx)
	githubID, githubIDOk := GetGitHubID(ctx)
	username, usernameOk := GetUsername(ctx)

	ok = userIDOk && githubIDOk && usernameOk
	return userID, githubID, username, ok
}

// GetOrganizationID retrieves the organization ID from context
func GetOrganizationID(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(OrganizationIDKey).(string)
	return orgID, ok
}
