package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	// Generate a random 32-byte (256-bit) secret key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}

	// Encode as base64 for easy use
	secret := base64.StdEncoding.EncodeToString(key)

	fmt.Printf("Generated JWT Secret (add to JWT_SECRET environment variable):\n%s\n", secret)
	fmt.Printf("# Example usage:\n# export JWT_SECRET=\"%s\"\n", secret)
}
