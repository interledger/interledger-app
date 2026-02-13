package ops_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/keys/ops"
	"gitlab.com/fynbos/env"
)

func TestCanAddAndSoftDeleteAPublicKey(t *testing.T) {
	env.SetEnv(t, "test")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil)
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)

	ks, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
	key := ks[0]
	require.Equal(t, walletID, key.WalletID)
	require.Equal(t, "database", key.Location)
	require.Equal(t, keys.NonCustodial.String(), key.Type.String())
	require.Equal(t, "My Key", key.Name)
	require.Equal(t, "123", key.KeyID)
	require.NotEmpty(t, key.CreatedAt)

	err = ops.DeletePublicKey(ctx, b, key.ID)
	require.NoError(t, err)

	ks, err = ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	assert.Empty(t, ks)

	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)

	ks, err = ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
	key = ks[0]
	require.Equal(t, walletID, key.WalletID)
	require.Equal(t, "database", key.Location)
	require.Equal(t, keys.NonCustodial.String(), key.Type.String())
	require.Equal(t, "My Key", key.Name)
	require.Equal(t, "123", key.KeyID)
}

func TestCantAddADuplicatePublicKey(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil)
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)

	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")

	require.ErrorIs(t, err, keys.ErrDuplicate)
}
