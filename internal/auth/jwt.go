package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrTokenExpired is returned when a token has expired
var ErrTokenExpired = errors.New("token expired")

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	GitHubID string `json:"github_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token
func GenerateToken(cfg JWTConfig, userID, githubID, username string) (string, error) {
	expirationTime := time.Now().Add(cfg.Expiration)
	claims := &Claims{
		UserID:   userID,
		GitHubID: githubID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken validates a JWT token and returns claims
// Returns ErrTokenExpired if the token is expired but otherwise valid
func ValidateToken(cfg JWTConfig, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		// Check if the error is due to token expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Even if expired, we can still extract claims for refresh
			if claims.UserID != "" && claims.GitHubID != "" {
				return claims, ErrTokenExpired
			}
		}
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
