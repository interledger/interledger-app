package storage

import (
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedTestUsers(t *testing.T) {
	store := NewMemoryStorage()
	err := SeedTestUsers(store)
	require.NoError(t, err)

	// User 1 exists with correct fields
	u1, err := store.GetUser(consts.TestUser1ID)
	require.NoError(t, err)
	assert.Equal(t, consts.TestUser1Email, u1.Email)
	assert.True(t, u1.Activated)
	assert.True(t, u1.Managed)
	assert.Equal(t, consts.KYCStateActionRequired, u1.KYCState)

	// User 2 exists
	u2, err := store.GetUser(consts.TestUser2ID)
	require.NoError(t, err)
	assert.Equal(t, consts.TestUser2Email, u2.Email)

	// Balances seeded
	usdBal, err := store.GetBalance(consts.TestUser1ID, "USD")
	require.NoError(t, err)
	assert.Equal(t, 10000.0, usdBal)

	eurBal, err := store.GetBalance(consts.TestUser2ID, "EUR")
	require.NoError(t, err)
	assert.Equal(t, 10000.0, eurBal)

	// Default organization created
	org, err := store.GetOrganization("default-org")
	require.NoError(t, err)
	assert.Equal(t, "sms", org.TwoFAType)
}

func TestSeedTestUsersWithOrgID(t *testing.T) {
	store := NewMemoryStorage()
	err := SeedTestUsersWithOrgID(store, "custom-org")
	require.NoError(t, err)

	org, err := store.GetOrganization("custom-org")
	require.NoError(t, err)
	assert.Equal(t, "custom-org", org.ID)
}

func TestSeedTestUsers_Idempotent(t *testing.T) {
	store := NewMemoryStorage()

	err := SeedTestUsers(store)
	require.NoError(t, err)

	// Seed again — should not error and should not change existing balances
	err = SeedTestUsers(store)
	require.NoError(t, err)

	// Balances remain unchanged on repeated seeding
	usdBal, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Equal(t, 10000.0, usdBal)

	eurBal, _ := store.GetBalance(consts.TestUser2ID, "EUR")
	assert.Equal(t, 10000.0, eurBal)
}
