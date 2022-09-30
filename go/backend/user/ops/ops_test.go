package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/backend/user/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()

	dbc := test_utils.MigrateCockroachDB(t, ctx)

	b := user_client.NewTestBackends(t, dbc, nil)

	userID := "c6874020-9d33-4678-a9ac-f623dc363cfb"

	// User does not have a signup
	_, err := ops.CreateWallet(ctx, b, userID, "test1")
	require.ErrorIs(t, err, user.ErrNoUserFound)

	// Create Signup
	_, err = dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)

	w, err := ops.CreateWallet(ctx, b, userID, "test1")
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	// Duplicate name should fail
	_, err = ops.CreateWallet(ctx, b, userID, "test1")
	require.ErrorIs(t, err, user.ErrDuplicateWallet)
}

func TestUserForContext(t *testing.T) {
	ctx := context.Background()

	_, err := ops.UserForContext(ctx)
	require.ErrorIs(t, err, user.ErrNoUserFound)

	ctx = context.WithValue(ctx, user.UserCtxKey("user"), &user.User{
		ID:    "1235",
		Email: "test@fynbos.dev",
	})

	u, err := ops.UserForContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, "1235")
	assert.Equal(t, u.Email, "test@fynbos.dev")
}

func TestWalletForContext(t *testing.T) {
	ctx := context.Background()

	_, err := ops.WalletForContext(ctx)
	require.ErrorIs(t, err, user.ErrNoWalletFound)

	ctx = context.WithValue(ctx, user.WalletCtxKey("wallet"), &user.Wallet{
		ID:   "1235",
		Name: "Default name",
	})

	w, err := ops.WalletForContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, w.ID, "1235")
	assert.Equal(t, w.Name, "Default name")
}

func TestListWallets(t *testing.T) {
	ctx := context.Background()

	dbc := test_utils.MigrateCockroachDB(t, ctx)

	b := user_client.NewTestBackends(t, dbc, nil)

	userID := "80629e7b-276b-4e38-82d5-8f73ef8c3806"
	// Create Signup
	_, err := dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)

	w, err := ops.CreateWallet(ctx, b, userID, "test1")
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	w, err = ops.CreateWallet(ctx, b, userID, "")
	require.NoError(t, err)
	assert.Equal(t, "default", w.Name)

	wallets, err := ops.ListWallets(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, wallets, 2)
}

func TestGetWallet(t *testing.T) {
	ctx := context.Background()
	dbc := test_utils.MigrateCockroachDB(t, ctx)
	b := user_client.NewTestBackends(t, dbc, nil)
	userID := uuid.NewString()
	// Create Signup
	_, err := dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)
	w, err := ops.CreateWallet(ctx, b, userID, "default")
	require.NoError(t, err)

	wallet, err := ops.GetWallet(ctx, b, userID, w.ID)

	require.NoError(t, err)
	require.Equal(t, w.ID, wallet.ID)
	require.Equal(t, w.Name, wallet.Name)
}
