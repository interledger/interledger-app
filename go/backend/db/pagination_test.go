package db_test

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/db"
	adminpb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"github.com/stretchr/testify/require"
)

func TestFromListWalletsPB(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		req := &adminpb.ListWalletsRequest{PageSize: 10}
		got := db.FromListWalletsPB(req)

		require.Equal(t, 10, got.PageSize)
		require.Equal(t, db.WalletFilter{}, got.Filter)
	})

	t.Run("populated subset", func(t *testing.T) {
		req := &adminpb.ListWalletsRequest{
			PageSize: 10,
			Filter: &adminpb.WalletSearchFilter{
				LastName:   "doe",
				ProviderId: "prov-123",
			},
		}
		got := db.FromListWalletsPB(req)

		require.Equal(t, db.WalletFilter{
			LastName:   "doe",
			ProviderID: "prov-123",
		}, got.Filter)
	})

	t.Run("whitespace-only values treated as empty", func(t *testing.T) {
		req := &adminpb.ListWalletsRequest{
			PageSize: 10,
			Filter: &adminpb.WalletSearchFilter{
				FirstName:     "   ",
				LastName:      "\t",
				WalletAddress: "  wallet.example  ",
				ProviderId:    "",
			},
		}
		got := db.FromListWalletsPB(req)

		require.Equal(t, db.WalletFilter{
			WalletAddress: "wallet.example",
		}, got.Filter)
	})

	t.Run("page-size cap still applied", func(t *testing.T) {
		require.Equal(t, 50, db.FromListWalletsPB(&adminpb.ListWalletsRequest{PageSize: 0}).PageSize)
		require.Equal(t, 50, db.FromListWalletsPB(&adminpb.ListWalletsRequest{PageSize: -1}).PageSize)
		require.Equal(t, 50, db.FromListWalletsPB(&adminpb.ListWalletsRequest{PageSize: 500}).PageSize)
		require.Equal(t, 25, db.FromListWalletsPB(&adminpb.ListWalletsRequest{PageSize: 25}).PageSize)
	})

	t.Run("pageToken and search passed through", func(t *testing.T) {
		req := &adminpb.ListWalletsRequest{
			PageSize:  10,
			PageToken: strPtr("tok-1"),
			Search:    strPtr("legacy-term"),
		}
		got := db.FromListWalletsPB(req)

		require.Equal(t, "tok-1", got.PageToken)
		require.Equal(t, "legacy-term", got.Search)
	})
}

func strPtr(s string) *string { return &s }
