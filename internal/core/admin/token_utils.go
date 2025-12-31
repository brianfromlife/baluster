package admin

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateTokenValue generates a random token value
func GenerateTokenValue() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
