package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/identities/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/env"
)

func TestAdd(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	w, err := userClient.CreateNewWallet(ctx, uuid.NewString(), "test")
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	_, err = ops.Add(ctx, b, identities.AddArgs{
		WalletID: w.ID,
		Platform: identities.PlatformTwitter,
		Handle:   "@king_cold",
		Public:   true,
	})
	require.NoError(t, err)

	// Not publicly visible
	_, err = ops.Add(ctx, b, identities.AddArgs{
		WalletID: w.ID,
		Platform: identities.PlatformTwitter,
		Handle:   "@cooler",
		Public:   false,
	})
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	w, err := userClient.CreateNewWallet(ctx, uuid.NewString(), "test")
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	_, err = ops.Add(ctx, b, identities.AddArgs{
		WalletID: w.ID,
		Platform: identities.PlatformTwitter,
		Handle:   "@king_cold",
		Public:   true,
	})
	require.NoError(t, err)

	// Not publicly visible
	_, err = ops.Add(ctx, b, identities.AddArgs{
		WalletID: w.ID,
		Platform: identities.PlatformTwitter,
		Handle:   "@cooler",
		Public:   false,
	})
	require.NoError(t, err)

	il, err := ops.List(ctx, b, w.ID)
	require.NoError(t, err)

	assert.Len(t, il, 2)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)
	assert.Equal(t, identities.PlatformTwitter, il[1].Platform)

	il, err = ops.ListPublic(ctx, b, w.ID)
	require.NoError(t, err)

	assert.Len(t, il, 1)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)
	assert.Equal(t, "@king_cold", il[0].Handle)
	assert.Equal(t, "", il[0].VerificationProof)
	assert.Equal(t, w.ID, il[0].WalletID)
	assert.Equal(t, identities.StateUnverified, il[0].State)
	assert.True(t, il[0].Public)
}
