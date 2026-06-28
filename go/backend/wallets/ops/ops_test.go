package ops_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/db"
	keys_mock "github.com/interledger/interledger-app/go/backend/keys/client/mock"
	users_mock "github.com/interledger/interledger-app/go/backend/user/client/mock"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/backend/wallets/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ensureTestDBURL(t *testing.T) {
	// Default to the local docker-compose credentials when DB_URL is unset.
	if os.Getenv("DB_URL") == "" {
		t.Setenv("DB_URL", "postgres://postgres:password@127.0.0.1:55432/%s?sslmode=disable")
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
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			_, errs[idx] = ops.Create(ctx, b, wallets.CreateArgs{UserID: userID})
		}(i)
	}

	close(start)
	wg.Wait()

	// Exactly one create wins; the rest are rejected as duplicates,
	// and the user is left with a single wallet.
	created := 0
	for _, err := range errs {
		if err == nil {
			created++
		} else {
			require.ErrorIs(t, err, wallets.ErrDuplicateWallet)
		}
	}
	require.Equal(t, 1, created)

	createdWallets, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, createdWallets, 1)
}

// A user may only ever have a single wallet (enforced by UNIQUE(user_id) on user_wallets).
// A second, named create is rejected and the user keeps their original wallet
func TestOneWalletPerUser(t *testing.T) {
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

	// A second wallet for the same user is rejected.
	_, err = ops.Create(ctx, b, wallets.CreateArgs{
		UserID:  userID,
		Name:    "za-wallet",
		Country: country.ZA,
	})
	require.ErrorIs(t, err, wallets.ErrDuplicateWallet)

	// The user still has exactly their original wallet.
	list, err := ops.List(ctx, b, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, usWallet.ID, list[0].ID)
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

func TestListAllSearch(t *testing.T) {
	ctx := context.Background()

	ensureTestDBURL(t)
	dbc := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	km := keys_mock.NewMockClient(ctrl)
	km.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b := ops.NewTestBackends(t, dbc, km, users_mock.NewMock())

	// ListAll is a global admin search; one wallet per user, so use distinct users.
	alpha, err := ops.Create(ctx, b, wallets.CreateArgs{UserID: uuid.NewString(), Name: "alpha"})
	require.NoError(t, err)

	_, err = ops.Create(ctx, b, wallets.CreateArgs{UserID: uuid.NewString(), Name: "beta", Country: country.ZA})
	require.NoError(t, err)

	_, err = ops.Create(ctx, b, wallets.CreateArgs{UserID: uuid.NewString(), Name: "gamma", Country: country.GB})
	require.NoError(t, err)

	t.Run("no search returns all wallets", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 50})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("search by exact name", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 50, Search: "alpha"})
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "alpha", result[0].Name)
	})

	t.Run("search by partial name case-insensitive", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 50, Search: "ALph"})
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "alpha", result[0].Name)
	})

	t.Run("search by wallet id", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 50, Search: alpha.ID})
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, alpha.ID, result[0].ID)
	})

	t.Run("search with no matches returns empty", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 50, Search: "zzznomatch"})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("search respects pagination", func(t *testing.T) {
		result, err := ops.ListAll(ctx, b, db.Pagination{PageSize: 1, Search: "a"})
		require.NoError(t, err)
		// pageSize+1 fetched — exactly 1 returned (the extra one is stripped by the handler, but here we test the raw DB layer)
		require.LessOrEqual(t, len(result), 2)
	})
}
