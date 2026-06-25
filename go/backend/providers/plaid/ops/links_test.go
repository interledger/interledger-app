package ops_test

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/db"
	ops "gitlab.com/fynbos/backend/providers/plaid/ops"
)

// linkTestBackends is the minimal duck-typed LinkBackends the plaid_links ops need.
type linkTestBackends struct {
	db        *sqlx.DB
	validator *validator.Validate
}

func (b linkTestBackends) DB() *sqlx.DB                   { return b.db }
func (b linkTestBackends) Validator() *validator.Validate { return b.validator }

// seedLinkedAccount inserts a minimal linked_accounts row so the plaid_links FK
// is satisfied, returning its id.
func seedLinkedAccount(t *testing.T, ctx context.Context, b linkTestBackends, walletID string) string {
	t.Helper()
	id := uuid.NewString()
	_, err := b.db.ExecContext(ctx,
		`INSERT INTO linked_accounts (id, wallet_id, name, mask, provider, provider_id, type)
		 VALUES ($1, $2, 'Plaid bank', '1234', 'pti', $3, 'bank');`,
		id, walletID, uuid.NewString(),
	)
	require.NoError(t, err)
	return id
}

func TestPlaidLinksLifecycle(t *testing.T) {
	ctx := context.Background()
	b := linkTestBackends{db: db.MigrateTestDB(t, ctx), validator: validator.New()}

	walletID := uuid.NewString()
	laID := seedLinkedAccount(t, ctx, b, walletID)
	const plaidAccountID = "plaid_account_abc"

	// Create
	link, err := ops.CreateLink(ctx, b, &ops.CreateLinkArgs{
		LinkedAccountID: laID,
		WalletID:        walletID,
		PlaidAccountID:  plaidAccountID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, link.ID)
	require.Equal(t, laID, link.LinkedAccountID)

	// Get
	got, err := ops.GetLinkByPlaidAccountID(ctx, b, walletID, plaidAccountID)
	require.NoError(t, err)
	require.Equal(t, link.ID, got.ID)

	// List
	ids, err := ops.ListPlaidAccountIDsByWallet(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, []string{plaidAccountID}, ids)

	// Partial unique index rejects a live duplicate (same wallet + plaid_account_id).
	laID2 := seedLinkedAccount(t, ctx, b, walletID)
	_, err = ops.CreateLink(ctx, b, &ops.CreateLinkArgs{
		LinkedAccountID: laID2,
		WalletID:        walletID,
		PlaidAccountID:  plaidAccountID,
	})
	require.ErrorIs(t, err, ops.ErrLinkInternal)

	// Soft delete frees the dedupe slot.
	require.NoError(t, ops.SoftDeleteLinkByLinkedAccountID(ctx, b, laID))

	_, err = ops.GetLinkByPlaidAccountID(ctx, b, walletID, plaidAccountID)
	require.ErrorIs(t, err, ops.ErrLinkNotFound)

	ids, err = ops.ListPlaidAccountIDsByWallet(ctx, b, walletID)
	require.NoError(t, err)
	require.Empty(t, ids)

	// Re-link the same (wallet, plaid_account_id) now succeeds.
	relinked, err := ops.CreateLink(ctx, b, &ops.CreateLinkArgs{
		LinkedAccountID: laID2,
		WalletID:        walletID,
		PlaidAccountID:  plaidAccountID,
	})
	require.NoError(t, err)
	require.NotEqual(t, link.ID, relinked.ID)

	// Invalid args are rejected before touching the DB.
	_, err = ops.CreateLink(ctx, b, &ops.CreateLinkArgs{WalletID: walletID, PlaidAccountID: plaidAccountID})
	require.ErrorIs(t, err, ops.ErrLinkInvalid)
}
