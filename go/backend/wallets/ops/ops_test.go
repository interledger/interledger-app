package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/wallets"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/backend/wallets/ops"
	"gotest.tools/assert"
)

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()

	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := user_client.NewTestBackends(t, dbc, nil, km)

	userID := "c6874020-9d33-4678-a9ac-f623dc363cfb"

	w, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
		Addresses: []wallets.Address{
			wallets.Address{},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	// Duplicate name should fail
	_, err = ops.CreateWallet(ctx, b, user.CreateWalletArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.ErrorIs(t, err, user.ErrDuplicateWallet)
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

	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := user_client.NewTestBackends(t, dbc, nil, km)

	userID := "80629e7b-276b-4e38-82d5-8f73ef8c3806"

	w, err := ops.CreateWallet(ctx, b, user.CreateWalletArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	w, err = ops.CreateWallet(ctx, b, user.CreateWalletArgs{
		UserID: userID,
		Name:   "",
	})
	require.NoError(t, err)
	assert.Equal(t, "default", w.Name)

	wallets, err := ops.ListWallets(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, wallets, 2)
}

func TestGetWallet(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := user_client.NewTestBackends(t, dbc, nil, km)
	userID := uuid.NewString()
	w, err := ops.CreateWallet(ctx, b, user.CreateWalletArgs{
		UserID: userID,
		Name:   "default",
	})
	require.NoError(t, err)

	wallet, err := ops.GetWallet(ctx, b, w.ID)

	require.NoError(t, err)
	require.Equal(t, w.ID, wallet.ID)
	require.Equal(t, w.Name, wallet.Name)
}

func TestSetWalletName(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := user_client.NewTestBackends(t, dbc, nil, km)
	userID := uuid.NewString()
	w, err := ops.CreateWallet(ctx, b, user.CreateWalletArgs{
		UserID: userID,
		Name:   "default",
	})
	require.NoError(t, err)
	require.Equal(t, "default", w.Name)

	err = ops.SetWalletName(ctx, b, w.ID, "Harry Potter")
	require.NoError(t, err)

	w, err = ops.GetWallet(ctx, b, w.ID)
	require.NoError(t, err)
	require.Equal(t, w.ID, w.ID)
	require.Equal(t, w.Name, "Harry Potter")
}
