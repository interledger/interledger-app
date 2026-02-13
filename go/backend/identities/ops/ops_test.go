package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/identities/ops"
	"gitlab.com/fynbos/env"
)

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

	res, err := ops.Search(ctx, b, uuid.NewString(), "notking")
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
