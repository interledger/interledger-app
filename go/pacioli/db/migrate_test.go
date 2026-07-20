package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// verifies the one-time migration backfill
// copies id into transaction_id for rows created before the column existed.
func TestBackfillTransferTransactionIDs(t *testing.T) {
	ctx := context.Background()
	connString, conn := MigrateTestDB(t, ctx)

	_, err := conn.ExecContext(ctx,
		"INSERT INTO ledgers (id, asset, scale) VALUES ($1, $2, $3)", 1, "USD", 2)
	require.NoError(t, err)

	debitID := uuid.NewString()
	creditID := uuid.NewString()
	for _, id := range []string{debitID, creditID} {
		_, err := conn.ExecContext(ctx,
			"INSERT INTO ledger_accounts (id, ledger_id, code) VALUES ($1, $2, $3)", id, 1, 1)
		require.NoError(t, err)
	}

	// transaction_id is left NULL, as it is for rows written before the backfill runs.
	transferID := uuid.NewString()
	_, err = conn.ExecContext(ctx,
		"INSERT INTO ledger_transfers (id, ledger_id, code, debit_account_id, credit_account_id, amount, state) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7)", transferID, 1, 1, debitID, creditID, 100, 4)
	require.NoError(t, err)

	require.NoError(t, backfillTransferTransactionIDs(ctx, connString))

	var txnID string
	require.NoError(t, conn.GetContext(ctx, &txnID,
		"SELECT transaction_id FROM ledger_transfers WHERE id=$1", transferID))
	require.Equal(t, transferID, txnID)

	// the backfill is idempotent: a second run must not change already-populated rows.
	require.NoError(t, backfillTransferTransactionIDs(ctx, connString))
	require.NoError(t, conn.GetContext(ctx, &txnID,
		"SELECT transaction_id FROM ledger_transfers WHERE id=$1", transferID))
	require.Equal(t, transferID, txnID)
}
