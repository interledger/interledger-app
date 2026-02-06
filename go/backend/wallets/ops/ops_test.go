package ops_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/backend/wallets/ops"
)

func ensureTestDBURL(t *testing.T) {
	// Default to the local docker-compose credentials when DB_URL is unset.
	if os.Getenv("DB_URL") == "" {
		if os.Getenv("CI") != "" {
			t.Fatalf("DB_URL must be set in CI to run database-backed tests")
		}
		t.Setenv("DB_URL", "postgres://postgres:password@0.0.0.0:5432/%s?sslmode=disable")
	}
}

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()

	ensureTestDBURL(t)
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
	assert.Equal(t, country.US, w.Country)

	// Duplicate name should fail
	_, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.ErrorIs(t, err, wallets.ErrDuplicateWallet)

	zaWalletID, zaUserID := uuid.NewString(), uuid.NewString()
	zaWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		ID:      zaWalletID,
		UserID:  zaUserID,
		Name:    "ZA Wallet",
		Country: country.ZA,
	})
	require.NoError(t, err)
	assert.Equal(t, country.ZA, zaWallet.Country)
}

func TestWalletForContext(t *testing.T) {
	ctx := context.Background()

	_, err := ops.WalletForContext(ctx)
	require.ErrorIs(t, err, wallets.ErrNoWalletFound)

	ctx = context.WithValue(ctx, wallets.CtxKey, &wallets.Wallet{
		ID:      "1235",
		Name:    "Default name",
		Country: country.US,
	})

	w, err := ops.WalletForContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, w.ID, "1235")
	assert.Equal(t, w.Name, "Default name")
	assert.Equal(t, country.US, w.Country)
}

func TestListWalletsSingle(t *testing.T) {
	// Split from TestListWallets to validate listing with a single wallet only.
	ctx := context.Background()

	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	userID := "80629e7b-276b-4e38-82d5-8f73ef8c3806"

	usWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", usWallet.Name)
	assert.Equal(t, country.US, usWallet.Country)

	wallets, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, wallets, 1)
	assert.Equal(t, usWallet.ID, wallets[0].ID)
	assert.Equal(t, country.US, wallets[0].Country)
}

func TestCreateDefaultWalletIsIdempotent(t *testing.T) {
	// Split from TestListWallets to validate default wallet creation is idempotent.
	ctx := context.Background()

	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	userID := "9fb3f58b-b0d6-4c5d-81b4-620c0d45f6b4"

	usWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)

	defaultWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID:  userID,
		Name:    "",
		Country: country.US,
	})
	require.NoError(t, err)
	assert.Equal(t, usWallet.ID, defaultWallet.ID)
	assert.Equal(t, usWallet.Name, defaultWallet.Name)
	assert.Equal(t, usWallet.Country, defaultWallet.Country)

	_, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID:  userID,
		Name:    "",
		Country: country.ZA,
	})
	require.ErrorIs(t, err, wallets.ErrWalletConflict)

	wa, err := wallets.ParseAddress("https://ilp.link/ladidaplah")
	require.NoError(t, err)
	_, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID:    userID,
		Name:      "",
		Country:   country.US,
		Addresses: []wallets.Address{wa},
	})
	require.ErrorIs(t, err, wallets.ErrWalletConflict)
}

func TestCreateDefaultWalletConcurrent(t *testing.T) {
	ctx := context.Background()

	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	userID := uuid.NewString()

	const workers = 10
	start := make(chan struct{})
	ids := make([]string, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			wallet, err := ops.Create(ctx, b, wallets.CreateArgs{
				UserID: userID,
				Name:   "",
			})
			errs[idx] = err
			if wallet != nil {
				ids[idx] = wallet.ID
			}
		}(i)
	}

	close(start)
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}

	firstID := ids[0]
	require.NotEmpty(t, firstID)
	for _, id := range ids {
		assert.Equal(t, firstID, id)
	}

	createdWallets, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, createdWallets, 1)
	assert.Equal(t, firstID, createdWallets[0].ID)
}

func TestListWalletsMultiple(t *testing.T) {
	// Split from TestListWallets to validate listing when multiple wallets exist.
	ctx := context.Background()

	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	userID := "74a84e5f-f95d-4bb8-9f56-f1e4a8a32d07"

	usWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID: userID,
		Name:   "test1",
	})
	require.NoError(t, err)
	assert.Equal(t, "test1", usWallet.Name)
	assert.Equal(t, country.US, usWallet.Country)

	zaWallet, err := ops.Create(ctx, b, wallets.CreateArgs{
		UserID:  userID,
		Name:    "za-wallet",
		Country: country.ZA,
	})
	require.NoError(t, err)
	assert.Equal(t, "za-wallet", zaWallet.Name)
	assert.Equal(t, country.ZA, zaWallet.Country)

	wallets, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, wallets, 2)
	for _, w := range wallets {
		if w.ID == usWallet.ID {
			assert.Equal(t, country.US, w.Country)
		} else {
			assert.Equal(t, country.ZA, w.Country)
		}
	}
}

func TestGetWallet(t *testing.T) {
	ctx := context.Background()
	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())
	wa, err := wallets.ParseAddress("https://ilp.link/ladidaplah")
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
	assert.Equal(t, w.Country, wallet.Country)
}

func TestSetWalletName(t *testing.T) {
	t.Skip("SKIPPING BROKEN TEST TODO FIX THIS")

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

func TestAddAddress(t *testing.T) {
	t.Skip("SKIPPING BROKEN TEST TODO FIX THIS")
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	uc := users_mock.NewMock()
	b := ops.NewTestBackends(t, db, km, uc)

	cases := []struct {
		name      string
		url       string
		duplicate bool
		err       error
		errMsg    string
	}{
		{
			name: "success",
			url:  "https://ilp.link/abcd1",
			err:  nil,
		},
		{
			name: "invalid_url",
			url:  "httpssss://ilp.link/creature",
			err:  wallets.ErrInvalidAddress,
		},
		{
			name:      "duplicate",
			url:       "https://ilp.link/abcd3",
			duplicate: true,
			err:       wallets.ErrAddressExists,
		},
		{
			name:   "regex_first_4_not_alpha",
			url:    "https://ilp.link/1234PayMe",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your first 3 characters must be letters",
		},
		{
			name:   "regex_contains_slash",
			url:    "https://ilp.link/PayMe/1234",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "regex_too_short",
			url:    "https://ilp.link/Pa",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address must be longer than 3 characters",
		},
		{
			name:   "regex_too_long",
			url:    "https://ilp.link/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address must be shorter than 30 characters",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			walletID := uuid.NewString()
			uc.WalletUser[walletID] = uuid.NewString()
			_, err := ops.Create(ctx, b, wallets.CreateArgs{
				ID:     walletID,
				UserID: uuid.NewString(),
				Name:   tc.name,
			})
			require.NoError(t, err)

			_, err = ops.AddAddress(ctx, b, walletID, tc.url)
			if tc.duplicate {
				require.NoError(t, err)
				_, err = ops.AddAddress(ctx, b, uuid.NewString(), tc.url)
			}

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				assert.True(t, strings.HasSuffix(err.Error(), tc.errMsg))
				assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
				return
			}

			require.NoError(t, err)

			// Lookup and validate
			w, err := ops.GetFromAddress(ctx, b, tc.url)
			require.NoError(t, err)
			assert.Equal(t, tc.name, w.Name)
			assert.Equal(t, walletID, w.ID)
			assert.Equal(t, tc.url, w.AddressString())
		})
	}
}
