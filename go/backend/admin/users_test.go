package admin

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUsersClient implements user.Client, overriding only the methods
// ListWallets exercises. Embedding a nil user.Client satisfies the rest of
// the interface (panicking if ever called, which none of these tests do).
type fakeUsersClient struct {
	user.Client
	findByEmailFn  func(ctx context.Context, email string) (string, error)
	findByPrefixFn func(ctx context.Context, term string) ([]string, error)
}

func (f *fakeUsersClient) FindWalletIDByEmail(ctx context.Context, email string) (string, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(ctx, email)
	}
	return "", nil
}

func (f *fakeUsersClient) FindWalletIDsByIdentifierPrefix(ctx context.Context, term string) ([]string, error) {
	if f.findByPrefixFn != nil {
		return f.findByPrefixFn(ctx, term)
	}
	return nil, nil
}

func (f *fakeUsersClient) ListUsers(_ context.Context, _ string) ([]user.User, error) {
	return nil, nil
}

// testBackends implements Backends, overriding only Users() and Wallets().
// Embedding a nil Backends satisfies the rest of the interface.
type testBackends struct {
	Backends
	users   user.Client
	wallets wallets.Client
}

func (b *testBackends) Users() user.Client      { return b.users }
func (b *testBackends) Wallets() wallets.Client { return b.wallets }

func TestListWallets_LegacyPathWhenFilterAbsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	wm := wallets_mock.NewMockClient(ctrl)
	wm.EXPECT().ListAll(gomock.Any(), gomock.Any()).Return([]wallets.Wallet{{ID: "w1", Name: "alice"}}, nil)

	uc := &fakeUsersClient{}
	s := &AdminRpcService{b: &testBackends{users: uc, wallets: wm}}

	resp, err := s.ListWallets(context.Background(), &adminv1.ListWalletsRequest{PageSize: 50, Search: strPtr("alice")})
	require.NoError(t, err)
	require.Len(t, resp.Wallets, 1)
	assert.Equal(t, "w1", resp.Wallets[0].WalletID)
}

func TestListWallets_LegacyEmailSearchResolvesNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	wm := wallets_mock.NewMockClient(ctrl)
	// Legacy email special-case with zero Kratos match must short-circuit
	// before ever calling ListAll.
	wm.EXPECT().ListAll(gomock.Any(), gomock.Any()).Times(0)

	uc := &fakeUsersClient{
		findByEmailFn: func(_ context.Context, _ string) (string, error) { return "", nil },
	}
	s := &AdminRpcService{b: &testBackends{users: uc, wallets: wm}}

	resp, err := s.ListWallets(context.Background(), &adminv1.ListWalletsRequest{PageSize: 50, Search: strPtr("nobody@example.com")})
	require.NoError(t, err)
	assert.Empty(t, resp.Wallets)
}

func TestListWallets_FilterEmailResolvesToNothing_ShortCircuits(t *testing.T) {
	ctrl := gomock.NewController(t)
	wm := wallets_mock.NewMockClient(ctrl)
	// Empty Kratos resolution must short-circuit — ListAll is never called.
	wm.EXPECT().ListAll(gomock.Any(), gomock.Any()).Times(0)

	uc := &fakeUsersClient{
		findByPrefixFn: func(_ context.Context, _ string) ([]string, error) { return nil, nil },
	}
	s := &AdminRpcService{b: &testBackends{users: uc, wallets: wm}}

	resp, err := s.ListWallets(context.Background(), &adminv1.ListWalletsRequest{
		PageSize: 50,
		Filter:   &adminv1.WalletSearchFilter{Email: "nonexistent-9f3a@x.invalid"},
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Wallets)
}

func TestListWallets_FilterEmailAndPhoneIntersect(t *testing.T) {
	walletA := uuid.NewString()
	walletB := uuid.NewString()

	ctrl := gomock.NewController(t)
	wm := wallets_mock.NewMockClient(ctrl)
	wm.EXPECT().ListAll(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, page db.Pagination) ([]wallets.Wallet, error) {
			require.Equal(t, []string{walletA}, page.Filter.WalletIDs)
			return []wallets.Wallet{{ID: walletA, Name: "matched"}}, nil
		})

	uc := &fakeUsersClient{
		findByPrefixFn: func(_ context.Context, term string) ([]string, error) {
			if term == "jane" {
				return []string{walletA, walletB}, nil
			}
			// phone resolves to walletA only — intersection must be walletA.
			return []string{walletA}, nil
		},
	}
	s := &AdminRpcService{b: &testBackends{users: uc, wallets: wm}}

	resp, err := s.ListWallets(context.Background(), &adminv1.ListWalletsRequest{
		PageSize: 50,
		Filter:   &adminv1.WalletSearchFilter{Email: "jane", PhoneNumber: "+1000"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Wallets, 1)
	assert.Equal(t, walletA, resp.Wallets[0].WalletID)
}

func TestListWallets_FilterFieldsPassedThroughToOpsLayer(t *testing.T) {
	ctrl := gomock.NewController(t)
	wm := wallets_mock.NewMockClient(ctrl)
	wm.EXPECT().ListAll(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, page db.Pagination) ([]wallets.Wallet, error) {
			assert.Equal(t, "jane", page.Filter.FirstName)
			assert.Equal(t, "doe", page.Filter.LastName)
			assert.Equal(t, "ilp.example", page.Filter.WalletAddress)
			assert.Equal(t, "prov-1", page.Filter.ProviderID)
			assert.Nil(t, page.Filter.WalletIDs)
			return nil, nil
		})

	uc := &fakeUsersClient{}
	s := &AdminRpcService{b: &testBackends{users: uc, wallets: wm}}

	_, err := s.ListWallets(context.Background(), &adminv1.ListWalletsRequest{
		PageSize: 50,
		Filter: &adminv1.WalletSearchFilter{
			FirstName:     "jane",
			LastName:      "doe",
			WalletAddress: "ilp.example",
			ProviderId:    "prov-1",
		},
	})
	require.NoError(t, err)
}

func strPtr(s string) *string { return &s }

func TestConvertUser(t *testing.T) {
	t.Run("maps all fields including first and last name", func(t *testing.T) {
		got := convertUser(user.User{
			ID:          "id-1",
			Email:       "jane@example.com",
			PhoneNumber: "+123456789",
			FirstName:   "Jane",
			LastName:    "Doe",
		})

		assert.Equal(t, "id-1", got.Id)
		assert.Equal(t, "jane@example.com", got.Email)
		assert.Equal(t, "+123456789", got.PhoneNumber)
		assert.Equal(t, "Jane", got.FirstName)
		assert.Equal(t, "Doe", got.LastName)
	})

	t.Run("empty names pass through as empty strings, no substitution", func(t *testing.T) {
		got := convertUser(user.User{ID: "id-2", Email: "no-name@example.com"})

		assert.Equal(t, "id-2", got.Id)
		assert.Empty(t, got.FirstName)
		assert.Empty(t, got.LastName)
		assert.Empty(t, got.PhoneNumber)
	})
}
