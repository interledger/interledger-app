package ops_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/keys/ops"
	vaultmock "gitlab.com/fynbos/backend/vault/mock"
	"gitlab.com/fynbos/env"
)

func TestGeneratePrivateAndListKeys(t *testing.T) {
	env.SetEnv(t, "test")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)

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
	require.Equal(t, "Interledger Managed", key.Name)
}

func TestCantGeneratePrivateDuplicateKeys(t *testing.T) {
	env.SetEnv(t, "test")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
	walletID := uuid.NewString()
	err := ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)
	ks, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)

	err = ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)

	ks, err = ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, ks, 1)
}

func TestCanAddAndSoftDeleteAPublicKey(t *testing.T) {
	env.SetEnv(t, "test")	
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
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
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)

	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")

	require.ErrorIs(t, err, keys.ErrDuplicate)
}

func TestCanSignAndVerifyCustodialKeys(t *testing.T) {
	env.SetEnv(t, "test")
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
	walletID := uuid.NewString()
	err := ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)
	keys, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	k := keys[0]
	message := []byte("Random message to sign")
	sig, err := ops.Sign(ctx, b, k.ID, k.WalletID, message)
	require.NoError(t, err)

	valid, err := ops.Verify(ctx, b, k.ID, k.WalletID, message, sig)
	require.NoError(t, err)

	require.True(t, valid)
}

func TestCantSignWithNonCustodialKeys(t *testing.T) {
	env.SetEnv(t, "test")	
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
	walletID := uuid.NewString()
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)
	keys, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	k := keys[0]

	message := []byte("Random message to sign")
	sig, err := ops.Sign(ctx, b, k.ID, k.WalletID, message)

	require.Error(t, err)
	require.Nil(t, sig)
}

func TestCanVerifyNonCustodialKeys(t *testing.T) {
	env.SetEnv(t, "test")	
	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, nil)
	walletID := uuid.NewString()
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	_, err = ops.AddPublicKey(ctx, b, walletID, base64.StdEncoding.EncodeToString(pubKey), "My Key", "123")
	require.NoError(t, err)
	keys, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	k := keys[0]

	message := []byte("Random message to sign")
	sig := ed25519.Sign(privKey, message)

	valid, err := ops.Verify(ctx, b, k.ID, k.WalletID, message, sig)
	require.NoError(t, err)

	require.True(t, valid)
}

/*
 * Test section for Vault based keys
 */

func TestGeneratePrivateVaultKey(t *testing.T) {
	t.Skip("Skipping test because we are not using vault anymore")
	env.SetEnv(t, "local")	
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	vc := vaultmock.NewMockClient(mockCtrl)
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), vc, nil)
	vc.EXPECT().CreateKey(gomock.Any())
	vc.EXPECT().GetPublicKey(gomock.Any()).Return("publicKey", nil)

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
	require.Equal(t, "vault", key.Location)
	require.Equal(t, keys.Custodial.String(), key.Type.String())
	require.Equal(t, "Interledger Managed", key.Name)
}

func TestCanSignAndVerifyCustodialKeysVault(t *testing.T) {
	t.Skip("Skipping because we are not using vault anymore")
	env.SetEnv(t, "local")	
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	vc := vaultmock.NewMockClient(mockCtrl)
	vc.EXPECT().CreateKey(gomock.Any())
	vc.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
	vc.EXPECT().GetPublicKey(gomock.Any()).Return("publicKey", nil)
	vc.EXPECT().Sign(gomock.Any(), gomock.Any()).Return([]byte("signature"), nil)
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), vc, nil)
	walletID := uuid.NewString()
	err := ops.GeneratePrivateKey(ctx, b, walletID)
	require.NoError(t, err)
	keys, err := ops.ListKeys(ctx, b, walletID)
	require.NoError(t, err)
	k := keys[0]

	message := []byte("Random message to sign")
	sig, err := ops.Sign(ctx, b, k.ID, k.WalletID, message)
	require.NoError(t, err)

	valid, err := ops.Verify(ctx, b, k.ID, k.WalletID, message, sig)
	require.NoError(t, err)

	require.True(t, valid)
}
