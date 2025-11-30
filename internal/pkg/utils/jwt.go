package utils

import (
	"os"
)

// GetJWTSecret retrieves the JWT secret from environment or returns default
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_change_in_prod"
	}
	return secret
}
