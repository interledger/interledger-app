package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	id := GenerateUUID()
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), id)

	// Two calls should produce different UUIDs
	assert.NotEqual(t, GenerateUUID(), GenerateUUID())
}

func TestGenerateMockXRPLAddress(t *testing.T) {
	addr := GenerateMockXRPLAddress()
	assert.True(t, len(addr) > 20, "address should be at least 20 chars")
	assert.Equal(t, byte('r'), addr[0], "XRPL address must start with 'r'")

	// Two calls should produce different addresses
	assert.NotEqual(t, GenerateMockXRPLAddress(), GenerateMockXRPLAddress())
}

func TestGenerateMockTransactionHash(t *testing.T) {
	hash := GenerateMockTransactionHash()
	assert.Len(t, hash, 64, "transaction hash should be 64 hex chars")
	assert.Regexp(t, regexp.MustCompile(`^[0-9A-F]{64}$`), hash)

	assert.NotEqual(t, GenerateMockTransactionHash(), GenerateMockTransactionHash())
}
