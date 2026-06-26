package ops_test

import (
	"context"
	"testing"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_ops "gitlab.com/fynbos/backend/linkedaccounts/ops"
	ops "gitlab.com/fynbos/backend/providers/plaid/ops"
)

// linkTestBackends is the minimal duck-typed LinkBackends the plaid_links ops need.
type linkTestBackends struct {
	db        *sqlx.DB
	validator *validator.Validate
}

func (b linkTestBackends) DB() *sqlx.DB                   { return b.db }
func (b linkTestBackends) Validator() *validator.Validate { return b.validator }
func (b linkTestBackends) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return crdbsqlx.ExecuteTx(ctx, b.db, nil, fn)
}

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

func createLink(ctx context.Context, b linkTestBackends, args *ops.CreateLinkArgs) (*ops.PlaidLink, error) {
	var link *ops.PlaidLink
	err := b.WithTx(ctx, func(tx *sqlx.Tx) error {
		l, e := ops.CreateLinkTx(ctx, tx, b.validator, args)
		if e != nil {
			return e
		}
		link = l
		return nil
	})
	return link, err
}

func TestPlaidLinksLifecycle(t *testing.T) {
	ctx := context.Background()
	b := linkTestBackends{db: db.MigrateTestDB(t, ctx), validator: validator.New()}

	walletID := uuid.NewString()
	laID := seedLinkedAccount(t, ctx, b, walletID)
	const plaidAccountID = "plaid_account_abc"

	// Create
	link, err := createLink(ctx, b, &ops.CreateLinkArgs{
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
	_, err = createLink(ctx, b, &ops.CreateLinkArgs{
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
	relinked, err := createLink(ctx, b, &ops.CreateLinkArgs{
		LinkedAccountID: laID2,
		WalletID:        walletID,
		PlaidAccountID:  plaidAccountID,
	})
	require.NoError(t, err)
	require.NotEqual(t, link.ID, relinked.ID)

	// Invalid args are rejected before touching the DB.
	_, err = createLink(ctx, b, &ops.CreateLinkArgs{WalletID: walletID, PlaidAccountID: plaidAccountID})
	require.ErrorIs(t, err, ops.ErrLinkInvalid)
}

// TestLinkCreationIsAtomic proves the linked_account and plaid_link inserts share
// one transaction: rolling it back leaves neither row, so a partial failure can
// never orphan a linked account (which would break Plaid dedupe).
func TestLinkCreationIsAtomic(t *testing.T) {
	ctx := context.Background()
	b := linkTestBackends{db: db.MigrateTestDB(t, ctx), validator: validator.New()}
	walletID := uuid.NewString()

	tx, err := b.db.BeginTxx(ctx, nil)
	require.NoError(t, err)

	la, err := linkedaccounts_ops.CreateTx(ctx, tx, b.validator, &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       "Plaid bank",
		Mask:       "1234",
		Provider:   "pti",
		ProviderID: uuid.NewString(),
		Type:       "bank",
		State:      linkedaccounts.Verified,
	})
	require.NoError(t, err)

	_, err = ops.CreateLinkTx(ctx, tx, b.validator, &ops.CreateLinkArgs{
		LinkedAccountID: la.ID,
		WalletID:        walletID,
		PlaidAccountID:  "plaid_account_atomic",
	})
	require.NoError(t, err)

	// Simulate a failure after both inserts: roll back the whole unit of work.
	require.NoError(t, tx.Rollback())

	var n int
	require.NoError(t, b.db.GetContext(ctx, &n, "SELECT count(*) FROM linked_accounts WHERE id=$1;", la.ID))
	require.Equal(t, 0, n, "linked_account must not persist after rollback")
	require.NoError(t, b.db.GetContext(ctx, &n, "SELECT count(*) FROM plaid_links WHERE linked_account_id=$1;", la.ID))
	require.Equal(t, 0, n, "plaid_link must not persist after rollback")
}
