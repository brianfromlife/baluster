package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"

	"github.com/brianfromlife/baluster/internal/auth"
	"github.com/brianfromlife/baluster/internal/core/admin"
	httputil "github.com/brianfromlife/baluster/internal/http"
	"github.com/brianfromlife/baluster/internal/storage"
)

type AuthHandler struct {
	githubOAuth *oauth2.Config
	stateCache  *auth.StateCache
	jwtConfig   auth.JWTConfig
	userRepo    *storage.UserRepository
	orgRepo     admin.OrganizationMemberLister
}

func NewAuthHandler(
	githubOAuth *oauth2.Config,
	stateCache *auth.StateCache,
	jwtConfig auth.JWTConfig,
	userRepo *storage.UserRepository,
	orgRepo admin.OrganizationMemberLister,
) *AuthHandler {
	return &AuthHandler{
		githubOAuth: githubOAuth,
		stateCache:  stateCache,
		jwtConfig:   jwtConfig,
		userRepo:    userRepo,
		orgRepo:     orgRepo,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/github", h.GitHubOAuth)
	r.Get("/github/callback", h.GitHubOAuthCallback)
	r.Post("/logout", h.Logout)
}

// GitHubOAuth initiates GitHub OAuth flow
func (h *AuthHandler) GitHubOAuth(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.githubOAuth.RedirectURL
	}

	input := &admin.GitHubOAuthInput{
		RedirectURI: redirectURI,
	}

	output, err := admin.GitHubOAuth(r.Context(), h.githubOAuth, h.stateCache, input)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.Success(w, http.StatusOK, map[string]interface{}{
		"auth_url": output.AuthURL,
		"state":    output.State,
	})
}

// GitHubOAuthCallback handles GitHub OAuth callback
func (h *AuthHandler) GitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("missing code or state"))
		return
	}

	input := &admin.GitHubOAuthCallbackInput{
		Code:  code,
		State: state,
	}

	output, err := admin.GitHubOAuthCallback(r.Context(), h.githubOAuth, h.stateCache, h.jwtConfig, h.userRepo, input)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.Success(w, http.StatusOK, map[string]any{
		"token": output.SessionToken,
		"user": map[string]any{
			"id":         output.User.ID,
			"github_id":  output.User.GitHubID,
			"username":   output.User.Username,
			"avatar_url": output.User.AvatarURL,
		},
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	githubID, ok := auth.GetGitHubID(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	input := &admin.GetCurrentUserInput{
		GithubID: githubID,
		UserID:   userID,
	}

	output, err := admin.GetCurrentUser(r.Context(), h.userRepo, h.orgRepo, input)
	if err != nil {
		fmt.Printf("error getting user: %v", err)
		httputil.Error(w, http.StatusNotFound, err)
		return
	}

	response := map[string]any{
		"id":         output.User.ID,
		"github_id":  output.User.GitHubID,
		"username":   output.User.Username,
		"avatar_url": output.User.AvatarURL,
	}

	if output.Organization != nil {
		response["organization"] = map[string]string{
			"name": output.Organization.Name,
			"id":   output.Organization.ID,
		}
	} else {
		response["organization"] = nil
	}

	httputil.Success(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	httputil.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}
