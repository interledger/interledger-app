package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
)

func TestUpsertInteracEmail(t *testing.T) {
	ctx := context.Background()

	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
	})

	walletID := uuid.NewString()
	email := "test@test.com"

	ne, err := ops.UpsertInteracEmail(ctx, b, walletID, email)

	require.NoError(t, err)
	assert.Equal(t, email, ne)

	ne, err = ops.GetInteracEmail(ctx, b, walletID)

	require.NoError(t, err)
	assert.Equal(t, email, ne)

	email = "test2@test2.com"

	ne, err = ops.UpsertInteracEmail(ctx, b, walletID, email)

	require.NoError(t, err)
	assert.Equal(t, email, ne)

	ne, err = ops.GetInteracEmail(ctx, b, walletID)

	require.NoError(t, err)
	assert.Equal(t, email, ne)
}

func TestGetChiWallet(t *testing.T) {
	ctx := context.Background()

	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
	})

	walletID := uuid.NewString()
	chiWallet := uuid.NewString()

	_, err := b.DB().ExecContext(ctx, "INSERT INTO chi_money_wallets (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", chiWallet, walletID)
	require.NoError(t, err)

	chiW, err := ops.GetChiWallet(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, chiWallet, chiW)

}
