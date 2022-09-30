package ops_test

import (
	"context"
	"strings"
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
		errMsg    string
	}{
		{
			name:      "success",
			url:       "http://fynbos.me/abcd1",
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
			err:       openpayments.ErrInvalidPointerURL,
		},
		{
			name:      "invalid_asset",
			url:       "http://fynbos.me/abcd2",
			alias:     "",
			assetCode: "FUzzY",
			scale:     2,
			err:       openpayments.ErrInvalidArgument,
		},
		{
			name:      "duplicate",
			url:       "http://fynbos.me/abcd3",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			duplicate: true,
			err:       openpayments.ErrPaymentPointerExists,
		},
		{
			name:      "regex_first_4_not_alpha",
			url:       "http://fynbos.me/1234PayMe",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your first 4 characters must be letters",
		},
		{
			name:      "regex_contains_slash",
			url:       "http://fynbos.me/PayMe/1234",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:      "regex_too_short",
			url:       "http://fynbos.me/Pay",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer must be longer than 4 characters",
		},
		{
			name:      "regex_too_long",
			url:       "http://fynbos.me/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			alias:     "",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer must be shorter than 30 characters",
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
				assert.True(t, strings.HasSuffix(err.Error(), tc.errMsg))
				assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
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

func TestExtractPaymentPointer(t *testing.T) {
	cases := []struct {
		name            string
		url             string
		expectedPointer string
		expectedSuffix  string
		err             error
	}{
		{
			name:            "no_suffix",
			url:             "http://fybos.me/asdf",
			expectedPointer: "http://fybos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name:            "no_suffix_trailing_slash",
			url:             "http://fybos.me/asdf/",
			expectedPointer: "http://fybos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name:            "incoming_payments_suffix",
			url:             "http://fybos.me/asdf/incoming-payments",
			expectedPointer: "http://fybos.me/asdf",
			expectedSuffix:  "incoming-payments",
			err:             nil,
		},
		{
			name:            "outgoing_payments_suffix",
			url:             "http://fybos.me/asdf/outgoing-payments",
			expectedPointer: "http://fybos.me/asdf",
			expectedSuffix:  "outgoing-payments",
			err:             nil,
		},
		{
			name: "invalid_url",
			url:  "ladida-blah",
			err:  openpayments.ErrInvalidPointerURL,
		},
		{
			name: "double_suffix",
			url:  "http://fybos.me/asdf/incoming-payments/outgoing-payments",
			err:  openpayments.ErrInvalidPointerURL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pp, suffix, err := ops.ExtractPaymentPointer(tc.url)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			assert.Equal(t, tc.expectedPointer, pp)
			assert.Equal(t, tc.expectedSuffix, suffix)
		})
	}
}

func TestListWalletPaymentPointers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL")

	wallet, err := userClient.CreateNewWallet(ctx, uuid.NewString(), "test")
	require.NoError(t, err)

	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "http://fynbos.me/payp1",
		WalletID:   wallet.ID,
		Alias:      "Alias1",
		Asset:      "USD",
		AssetScale: 2,
	})
	require.NoError(t, err)
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "http://fynbos.me/payp2",
		WalletID:   wallet.ID,
		Alias:      "Alias2",
		Asset:      "ZAR",
		AssetScale: 2,
	})
	require.NoError(t, err)

	// List and validate
	pp, err := ops.ListWalletPaymentPointers(ctx, b, wallet.ID)
	require.NoError(t, err)
	require.Len(t, pp, 2)
	for _, p := range pp {
		if p.URL == "http://fynbos.me/payp1" {
			assert.Equal(t, "Alias1", p.Alias)
			assert.Equal(t, "USD", p.Asset)
			assert.Equal(t, 2, p.AssetScale)
		} else {
			assert.Equal(t, "Alias2", p.Alias)
			assert.Equal(t, "ZAR", p.Asset)
			assert.Equal(t, 2, p.AssetScale)
		}
	}
}
