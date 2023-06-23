package ops_test

import (
	"context"
	"testing"

	users_mock "gitlab.com/fynbos/backend/user/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/backend/wallets/ops"
	"gotest.tools/assert"
)

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()

	dbc := db.MigrateTestDB(t, ctx)

	userID := "c6874020-9d33-4678-a9ac-f623dc363cfb"
	walletID := uuid.NewString()

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	um := users_mock.NewMock()
	um.WalletUser[walletID] = userID
	b := ops.NewTestBackends(t, dbc, km, um)

	w, err := ops.Create(ctx, b, wallets.CreateArgs{
		ID:     walletID,
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	// Duplicate name should fail
	_, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.ErrorIs(t, err, wallets.ErrDuplicateWallet)
}

func TestWalletForContext(t *testing.T) {
	ctx := context.Background()

	_, err := ops.WalletForContext(ctx)
	require.ErrorIs(t, err, wallets.ErrNoWalletFound)

	ctx = context.WithValue(ctx, wallets.CtxKey, &wallets.Wallet{
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
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	userID := "80629e7b-276b-4e38-82d5-8f73ef8c3806"

	w, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", w.Name)

	w, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "",
	})
	require.NoError(t, err)
	assert.Equal(t, "default", w.Name)

	wallets, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, wallets, 2)
}

func TestGetWallet(t *testing.T) {
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())
	wa, err := wallets.ParseAddress("https://fynbos.me/ladidaplah")
	require.NoError(t, err)
	userID := uuid.NewString()
	w, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID:    userID,
		Name:      "default",
		Addresses: []wallets.Address{wa},
	})
	require.NoError(t, err)

	wallet, err := ops.Get(ctx, b, w.ID)

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
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())
	userID := uuid.NewString()
	w, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "default",
	})
	require.NoError(t, err)
	require.Equal(t, "default", w.Name)

	_, err = ops.SetWalletName(ctx, b, w.ID, "Harry Potter")
	require.NoError(t, err)

	w, err = ops.Get(ctx, b, w.ID)
	require.NoError(t, err)
	require.Equal(t, w.ID, w.ID)
	require.Equal(t, w.Name, "Harry Potter")
}
