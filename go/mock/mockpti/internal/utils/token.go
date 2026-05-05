package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateToken generates a random hex token.
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTokenExpiresAt returns an expiration time 55 minutes from now.
func GenerateTokenExpiresAt() time.Time {
	return time.Now().Add(55 * time.Minute)
}
