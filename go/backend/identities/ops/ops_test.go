package ops_test

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/identities/ops"
	"github.com/interledger/interledger-app/go/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	env.SetEnv(t, "local")

	// Publicly visible
	_, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   uuid.NewString(),
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	env.SetEnv(t, "local")

	walletID := uuid.NewString()
	// Publicly visible
	pv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	il, err := ops.List(ctx, b, walletID)
	require.NoError(t, err)

	assert.Len(t, il, 1)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)

	il, err = ops.ListPublic(ctx, b, walletID)
	require.NoError(t, err)

	require.Len(t, il, 0)

	// Verify public identity
	_, err = db.ExecContext(ctx, "UPDATE identities SET state=$1 WHERE id=$2", identities.StateVerified, pv.ID)
	require.NoError(t, err)

	il, err = ops.ListPublic(ctx, b, walletID)
	require.NoError(t, err)

	require.Len(t, il, 1)
	assert.Equal(t, identities.PlatformTwitter, il[0].Platform)
	assert.Equal(t, "@king_cold", il[0].Identifier)
	assert.Equal(t, "", il[0].VerificationProof)
	assert.Equal(t, walletID, il[0].WalletID)
	assert.Equal(t, identities.StateVerified, il[0].State)
	assert.True(t, il[0].Public)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	walletID := uuid.NewString()

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	_, err = ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)

	err = ops.Delete(ctx, b, iv.ID, walletID)
	require.NoError(t, err)

	_, err = ops.Get(ctx, b, iv.ID)
	require.ErrorIs(t, err, identities.ErrNotFound)
}

func TestSetPublic(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	walletID := uuid.NewString()

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	id, err := ops.Get(ctx, b, iv.ID)
	require.NoError(t, err)
	assert.True(t, id.Public)

	id, err = ops.SetPublic(ctx, b, iv.ID, walletID, false)
	require.NoError(t, err)
	assert.False(t, id.Public)
}

func TestVerified(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	walletID := uuid.NewString()

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
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

func TestGetBySignatureHash(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	walletID := uuid.NewString()

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "@king_cold",
	})
	require.NoError(t, err)

	id, err := ops.GetBySignatureHash(ctx, b, iv.SignatureHash)
	require.NoError(t, err)
	assert.True(t, id.Public)
	assert.Equal(t, iv.ID, id.ID)
}

func TestGetByIdentifier(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	walletID := uuid.NewString()

	env.SetEnv(t, "local")

	// Publicly visible
	iv, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "king_cold",
	})
	require.NoError(t, err)

	id, err := ops.GetByIdentifier(ctx, b, iv.Identifier)
	require.NoError(t, err)
	assert.True(t, id.Public)
	assert.Equal(t, iv.ID, id.ID)

	_, err = ops.GetByIdentifier(ctx, b, "notfound")
	require.ErrorIs(t, err, identities.ErrNotFound)
}

func TestSearch(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := ops.NewTestBackends(t, db)

	env.SetEnv(t, "local")
	walletID := uuid.NewString()

	_, err := b.DB().ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", walletID, "warmer")
	require.NoError(t, err)

	b.DB().MustExecContext(ctx, "INSERT INTO linked_accounts (wallet_id, name, mask, provider, type, provider_id, state, can_receive) VALUES ($1, $2, $3, $4, $5, $6, $7,$8)",
		walletID, "warmer", "1234", "testing", "card", uuid.NewString(), linkedaccounts.Verified, true)

	b.DB().MustExecContext(ctx, "INSERT INTO wallet_addresses (wallet_id, url) VALUES ($1, $2)",
		walletID, env.OpenPaymentsURL()+"/notking")

	// add approved KYC
	b.DB().MustExecContext(ctx, "INSERT INTO  wallet_kyc_status (wallet_id, status, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())",
		walletID, kyc.StatusLevel2)
	// Publicly visible
	id, err := ops.Add(ctx, b, identities.AddArgs{
		WalletID:   walletID,
		Platform:   identities.PlatformTwitter,
		Identifier: "king_cold",
	})
	require.NoError(t, err)

	_, err = ops.SetPublic(ctx, b, id.ID, walletID, true)
	require.NoError(t, err)

	err = ops.UpdateState(ctx, b, id.ID, identities.StateVerified, "proof")
	require.NoError(t, err)

	res, err := ops.Search(ctx, b, uuid.NewString(), "cold")
	require.NoError(t, err)

	assert.Len(t, res, 1)
	assert.Len(t, res[0].SubResults, 1)
	assert.Equal(t, "wallet", res[0].IdentifierType)
	assert.Equal(t, "warmer", res[0].Identifier)
	assert.Equal(t, 0.5, res[0].Rank)
	assert.Equal(t, env.OpenPaymentsURL()+"/notking", res[0].WalletUrl)
	assert.Equal(t, string(identities.PlatformTwitter), res[0].SubResults[0].IdentifierType)
	assert.Equal(t, "king_cold", res[0].SubResults[0].Identifier)
	assert.Equal(t, 0.5, res[0].SubResults[0].Rank)
	assert.Equal(t, env.OpenPaymentsURL()+"/notking", res[0].SubResults[0].WalletUrl)

	// Can't search for yourself
	res, err = ops.Search(ctx, b, walletID, "cold")
	require.NoError(t, err)
	assert.Len(t, res, 0)

	// Search with a full twitter URL
	res, err = ops.Search(ctx, b, uuid.NewString(), "https://twitter.com/king_cold")
	require.NoError(t, err)
	assert.Len(t, res[0].SubResults, 1)
	assert.Equal(t, "wallet", res[0].IdentifierType)
	assert.Equal(t, "warmer", res[0].Identifier)
	assert.Equal(t, float64(1), res[0].Rank)
	assert.Equal(t, env.OpenPaymentsURL()+"/notking", res[0].WalletUrl)
	assert.Equal(t, string(identities.PlatformTwitter), res[0].SubResults[0].IdentifierType)
	assert.Equal(t, "king_cold", res[0].SubResults[0].Identifier)
	assert.Equal(t, float64(1), res[0].SubResults[0].Rank)
	assert.Equal(t, env.OpenPaymentsURL()+"/notking", res[0].SubResults[0].WalletUrl)

	// Search payment pointer
	res, err = ops.Search(ctx, b, uuid.NewString(), "notking")
	require.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, string("wallet"), res[0].IdentifierType)
	assert.Equal(t, "warmer", res[0].Identifier)
	assert.Equal(t, string("wallet_url"), res[0].SubResults[0].IdentifierType)
	assert.Equal(t, env.OpenPaymentsURL()+"/notking", res[0].SubResults[0].Identifier)

	// Now for a grouping, wallet and twitter name matches so group em
	_, err = b.DB().ExecContext(ctx, "UPDATE wallets SET name=$1 WHERE id = $2", "cold_iron", walletID)
	require.NoError(t, err)

	res, err = ops.Search(ctx, b, uuid.NewString(), "cold")
	require.NoError(t, err)

	require.Len(t, res, 1)
	assert.Equal(t, "wallet", res[0].IdentifierType)
	assert.Equal(t, "cold_iron", res[0].Identifier)
	assert.Equal(t, 0.5, res[0].Rank)
	require.Len(t, res[0].SubResults, 1)
	assert.Equal(t, string(identities.PlatformTwitter), res[0].SubResults[0].IdentifierType)
	assert.Equal(t, "king_cold", res[0].SubResults[0].Identifier)
	assert.Equal(t, float64(0.5), res[0].SubResults[0].Rank)

}
