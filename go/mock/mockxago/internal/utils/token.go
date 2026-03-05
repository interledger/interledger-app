package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateToken generates a random hex token
func GenerateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// GenerateTokenExpiresAt generates an expiration time (55 minutes from now)
func GenerateTokenExpiresAt() time.Time {
	return time.Now().Add(55 * time.Minute)
}
