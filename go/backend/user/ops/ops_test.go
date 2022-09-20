package ops_test

import (
	"context"
	"testing"

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

	w, err := ops.CreateWallet(ctx, b, userID, "test1")
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)
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
