package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreatePaymentPointer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL")

	cases := []struct {
		name      string
		url       string
		alias     string
		assetCode string
		scale     int
		duplicate bool
		err       error
	}{
		{
			name:      "success",
			url:       "http://fynbos.me/1",
			alias:     "Alias",
			assetCode: "USD",
			scale:     2,
			err:       nil,
		},
		{
			name:      "invalid_url",
			url:       "fynbos.me/creature",
			alias:     "",
			assetCode: "USD",
			scale:     2,
			err:       openpayments.ErrInvalidArgument,
		},
		{
			name:      "invalid_asset",
			url:       "http://fynbos.me/2",
			alias:     "",
			assetCode: "FUzzY",
			scale:     2,
			err:       openpayments.ErrInvalidArgument,
		},
		{
			name:      "duplicate",
			url:       "http://fynbos.me/3",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			duplicate: true,
			err:       openpayments.ErrPaymentPointerExists,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			wallet, err := userClient.CreateNewWallet(ctx, uuid.NewString(), "test")
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.url,
				WalletID:   wallet.ID,
				Alias:      tc.alias,
				Asset:      tc.assetCode,
				AssetScale: tc.scale,
			})
			if tc.duplicate {
				require.NoError(t, err)
				err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
					URL:        tc.url,
					WalletID:   wallet.ID,
					Alias:      tc.alias,
					Asset:      tc.assetCode,
					AssetScale: tc.scale,
				})
			}

			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)

			// Lookup and validate
			pp, err := ops.GetPaymentPointer(ctx, b, tc.url)
			require.NoError(t, err)
			assert.Equal(t, tc.alias, pp.Alias)
			assert.Equal(t, wallet.ID, pp.WalletID)
			assert.Equal(t, tc.url, pp.URL)
			assert.Equal(t, tc.assetCode, pp.Asset)
			assert.Equal(t, tc.scale, pp.AssetScale)
		})
	}
}
