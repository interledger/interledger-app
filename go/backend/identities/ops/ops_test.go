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
	"gitlab.com/fynbos/backend/user"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/env"
)

func TestAdd(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_mock.NewMock()

	w, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	_, err = ops.Add(ctx, b, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_mock.NewMock()

	w, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	pv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	il, err := ops.List(ctx, b, w.ID)
	require.NoError(t, err)

	assert.Len(t, il, 1)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)

	il, err = ops.ListPublic(ctx, b, w.ID)
	require.NoError(t, err)

	require.Len(t, il, 0)

	// Verify public identity
	_, err = db.ExecContext(ctx, "UPDATE identities SET state=$1 WHERE id=$2", identities.StateVerified, pv.ID)
	require.NoError(t, err)

	il, err = ops.ListPublic(ctx, b, w.ID)
	require.NoError(t, err)

	require.Len(t, il, 1)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)
	assert.Equal(t, "@king_cold", il[0].Identifier)
	assert.Equal(t, "", il[0].VerificationProof)
	assert.Equal(t, w.ID, il[0].WalletID)
	assert.Equal(t, identities.StateVerified, il[0].State)
	assert.True(t, il[0].Public)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_mock.NewMock()

	w, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	_, err = ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)

	err = ops.Delete(ctx, b, iv.ID, w.ID)
	require.NoError(t, err)

	_, err = ops.Get(ctx, b, iv.ID)
	require.ErrorIs(t, err, identities.ErrNotFound)
}

func TestSetPublic(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_mock.NewMock()

	w, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	id, err := ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)
	assert.True(t, id.Public)

	id, err = ops.SetPublic(ctx, b, iv.ID, w.ID, false)
	require.NoError(t, err)
	assert.False(t, id.Public)
}

func TestVerified(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	userClient := users_mock.NewMock()

	w, err := userClient.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   w.ID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	id, err := ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)
	assert.True(t, id.Public)
	assert.Equal(t, identities.StateUnverified, id.State)
	assert.False(t, id.VerifiedAt.Valid)

	err = ops.UpdateState(ctx, b, id.ID, identities.StateVerified, "")
	require.NoError(t, err)

	id, err = ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)
	assert.Equal(t, identities.StateVerified, id.State)
	assert.True(t, id.VerifiedAt.Valid)
}
