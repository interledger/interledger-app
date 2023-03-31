package ops_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/keys/ops"
	"gitlab.com/fynbos/env"
	"testing"
)

func TestGeneratePrivateAndListKeys(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))

	walletID := uuid.NewString()

	ks, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 0)

	err = ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)

	ks, err = ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
	key := ks[0]
	require.Equal(t, walletID, key.WalletID)
	require.Equal(t, "database", key.Location)
	require.Equal(t, keys.Custodial.String(), key.Type.String())
	require.Equal(t, "Fynbos Managed", key.Name)
}

func TestCantGeneratePrivateDuplicateKeys(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))
	walletID := uuid.NewString()
	err := ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)
	ks, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)

	err = ops.GeneratePrivateKey(ctx, b, walletID)

	ks, err = ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
}

func TestCanAddAPublicKey(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)

	err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key")

	ks, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
	key := ks[0]
	require.Equal(t, walletID, key.WalletID)
	require.Equal(t, "database", key.Location)
	require.Equal(t, keys.NonCustodial.String(), key.Type.String())
	require.Equal(t, "My Key", key.Name)
}

func TestCantAddADuplicatePublicKey(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)
	err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key")
	require.NoError(t, err)

	err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key")

	require.ErrorIs(t, err, keys.ErrDuplicate)
}
