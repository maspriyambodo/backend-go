package utils

import (
	"os"
	"sync"
)

var (
	jwtSecret     string
	jwtSecretOnce sync.Once
)

// GetJWTSecret retrieves the JWT secret from environment (cached after first call)
func GetJWTSecret() string {
	jwtSecretOnce.Do(func() {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default_secret_change_in_prod"
		}
	})
	return jwtSecret
}
