package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	// Generate multiple tokens
	token1 := GenerateToken()
	token2 := GenerateToken()

	// Verify tokens are non-empty
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)

	// Verify tokens are unique
	assert.NotEqual(t, token1, token2)

	// Verify token is hex encoded (64 chars for 32 bytes)
	assert.Len(t, token1, 64)
	assert.Len(t, token2, 64)

	// Verify token only contains hex characters
	assert.Regexp(t, "^[0-9a-f]+$", token1)
	assert.Regexp(t, "^[0-9a-f]+$", token2)
}

func TestGenerateTokenExpiresAt(t *testing.T) {
	now := time.Now()
	expiresAt := GenerateTokenExpiresAt()

	// Verify expiration is in the future
	assert.True(t, expiresAt.After(now))

	// Verify expiration is approximately 55 minutes from now (within 1 second tolerance)
	expectedExpiration := now.Add(55 * time.Minute)
	diff := expiresAt.Sub(expectedExpiration)
	assert.Less(t, diff, time.Second)
	assert.Greater(t, diff, -time.Second)
}
