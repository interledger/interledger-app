package ops_test

import (
	"context"
	"strings"
	"testing"
	"time"

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

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

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
			err:       openpayments.ErrInvalidPointerPath,
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

			userID := uuid.NewString()
			// Create Signup
			_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
			require.NoError(t, err)
			// Create Wallet
			wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
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
			url:             "http://fynbos.me/asdf",
			expectedPointer: "http://fynbos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name:            "no_suffix_trailing_slash",
			url:             "http://fynbos.me/asdf/",
			expectedPointer: "http://fynbos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name:            "incoming_payments_suffix",
			url:             "http://fynbos.me/asdf/incoming-payments",
			expectedPointer: "http://fynbos.me/asdf",
			expectedSuffix:  "incoming-payments",
			err:             nil,
		},
		{
			name:            "outgoing_payments_suffix",
			url:             "http://fynbos.me/asdf/outgoing-payments",
			expectedPointer: "http://fynbos.me/asdf",
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
			url:  "http://fynbos.me/asdf/incoming-payments/outgoing-payments",
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

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	userID := uuid.NewString()
	// Create Signup
	_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)

	wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
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

func TestValidatePaymentPointer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	cases := []struct {
		name   string
		url    string
		err    error
		errMsg string
		exists bool
	}{
		{
			name: "success",
			url:  "http://fynbos.me/abcd1",
			err:  nil,
		},
		{
			name: "invalid_url",
			url:  "fynbos.me/creature",
			err:  openpayments.ErrInvalidPointerPath,
		},
		{
			name:   "regex_first_4_not_alpha",
			url:    "http://fynbos.me/1234PayMe",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your first 4 characters must be letters",
		},
		{
			name:   "regex_contains_slash",
			url:    "http://fynbos.me/PayMe/1234",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "regex_too_short",
			url:    "http://fynbos.me/Pay",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer must be longer than 4 characters",
		},
		{
			name:   "regex_too_long",
			url:    "http://fynbos.me/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer must be shorter than 30 characters",
		},
		{
			name:   "contains_hash",
			url:    "http://fynbos.me/asdfn#asdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_space",
			url:    "http://fynbos.me/asdfn%20asdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_<",
			url:    "http://fynbos.me/asdfn%3Casdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "trailing_slash",
			url:    "http://fynbos.me/asdfn/",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "query_params",
			url:    "http://fynbos.me/asdfn?asdf3e=34334",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_escapes",
			url:    "http://fynbos.me/asdfn%2",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_dots",
			url:    "http://fynbos.me/abcdef..sasd",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			exists, err := ops.PaymentPointerExists(ctx, b, tc.url)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				assert.True(t, strings.HasSuffix(err.Error(), tc.errMsg))
				assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
				return
			}

			require.Equal(t, tc.exists, exists)
		})
	}
}

func TestPaymentPointerCaseSensitive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)
	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	userID := uuid.NewString()
	// Create Signup
	_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), userID)
	require.NoError(t, err)
	// Create Wallet
	wallet, err := userClient.CreateNewWallet(ctx, userID, "test")
	require.NoError(t, err)

	exists, err := ops.PaymentPointerExists(ctx, b, "https://fynbos.me/ValidPaymentPointer")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create payment pointer
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/ValidPaymentPointer",
		WalletID:   wallet.ID,
		Asset:      "ZAR",
		AssetScale: 2,
	})
	require.NoError(t, err)

	// Lookup with different case
	pp, err := ops.GetPaymentPointer(ctx, b, "https://fynBos.Me/validpaymentPointeR")
	require.NoError(t, err)
	assert.Equal(t, "https://fynbos.me/ValidPaymentPointer", pp.URL)

	// Exists with different case
	exists, err = ops.PaymentPointerExists(ctx, b, "https://fynbos.me/VALIDPAYMENTPOINTER")
	require.NoError(t, err)
	require.True(t, exists)

	// Create with different casing
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/VaLidPaymenTPoinTer",
		WalletID:   wallet.ID,
		Asset:      "ZAR",
		AssetScale: 2,
	})
	require.ErrorIs(t, err, openpayments.ErrPaymentPointerExists)
}

func TestFormattedPaymentPointer(t *testing.T) {
	cases := []struct {
		name              string
		url               string
		expectedFormatted string
		err               error
	}{
		{
			name:              "http",
			url:               "http://fynbos.me/asdf",
			expectedFormatted: "$fynbos.me/asdf",
			err:               nil,
		},
		{
			name:              "https",
			url:               "https://fynbos.me/asdf",
			expectedFormatted: "$fynbos.me/asdf",
			err:               nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := ops.FormattedPaymentPointer(tc.url)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			assert.Equal(t, tc.expectedFormatted, formatted)
		})
	}
}

func TestCreateQuote(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name      string
		args      openpayments.CreateQuoteArgs
		recvAsset string
		err       error
	}{
		{
			name: "success",
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "http://fynbos.me/paysend",
				ReceivePaymentPointer: "http://fynbos.me/payrecv",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
			},
		},
		{
			name:      "different assets",
			recvAsset: "ZAR",
			err:       openpayments.ErrInvalidArgument,
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "http://fynbos.me/paysend1",
				ReceivePaymentPointer: "http://fynbos.me/payrecv2",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
			},
		},
		{
			name: "expiry in the past",
			err:  openpayments.ErrInvalidArgument,
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "http://fynbos.me/paysend3",
				ReceivePaymentPointer: "http://fynbos.me/payrecv4",
				ExpiresAt:             time.Now().Add(time.Hour * -1),
				SendAmount: openpayments.Amount{
					Value:      100,
					Asset:      "ZAR",
					AssetScale: 2,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sendUserID := uuid.NewString()
			recvUserID := uuid.NewString()
			// Create Signups
			_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), sendUserID, uuid.NewString(), recvUserID)
			require.NoError(t, err)
			// Create Wallets
			sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
			require.NoError(t, err)
			recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.SendPaymentPointer,
				WalletID:   sendWallet.ID,
				Alias:      "Alias",
				Asset:      tc.args.SendAmount.Asset,
				AssetScale: tc.args.SendAmount.AssetScale,
			})
			require.NoError(t, err)
			recvAsset := tc.args.SendAmount.Asset
			if tc.recvAsset != "" {
				recvAsset = tc.recvAsset
			}
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.ReceivePaymentPointer,
				WalletID:   recvWallet.ID,
				Alias:      "Alias",
				Asset:      recvAsset,
				AssetScale: tc.args.SendAmount.AssetScale,
			})
			require.NoError(t, err)

			q, err := ops.CreateQuote(ctx, b, tc.args)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.False(t, q.ExpiresAt.IsZero())
			// We do this for clock drift
			assert.Equal(t, tc.args.ExpiresAt.Year(), q.ExpiresAt.Year())
			assert.Equal(t, tc.args.ExpiresAt.Month(), q.ExpiresAt.Month())
			assert.Equal(t, tc.args.ExpiresAt.Day(), q.ExpiresAt.Day())
			assert.Equal(t, tc.args.ExpiresAt.Hour(), q.ExpiresAt.Hour())
			assert.Equal(t, tc.args.ExpiresAt.Minute(), q.ExpiresAt.Minute())
			assert.Equal(t, tc.args.ExpiresAt.Second(), q.ExpiresAt.Second())
			assert.Equal(t, tc.args.SendPaymentPointer, q.PaymentPointer)
			assert.Equal(t, tc.args.SendAmount.Value, q.ReceiveAmount.Value)
			assert.Equal(t, tc.args.SendAmount.Asset, q.ReceiveAmount.Asset)
			assert.Equal(t, tc.args.SendAmount.AssetScale, q.ReceiveAmount.AssetScale)
			assert.Equal(t, tc.args.SendAmount.Value, q.SendAmount.Value)
			assert.Equal(t, tc.args.SendAmount.Asset, q.SendAmount.Asset)
			assert.Equal(t, tc.args.SendAmount.AssetScale, q.SendAmount.AssetScale)
		})
	}
}

func TestCreateIncomingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name    string
		ppAsset string
		args    openpayments.CreateIncomingPaymentArgs
		err     error
	}{
		{
			name: "success",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer: "http://fynbos.me/moneyplease",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "success no incoming amount",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer: "http://fynbos.me/moneyplease4",
				ExternalRef:    "external",
				ExpiresAt:      time.Now().Add(time.Hour),
			},
		},
		{
			name:    "different assets",
			ppAsset: "ZAR",
			err:     openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer: "http://fynbos.me/moneyplease2",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "past expiry",
			err:  openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer: "http://fynbos.me/moneyplease3",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour * -1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recvUserID := uuid.NewString()
			// Create Signups
			_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2)", uuid.NewString(), recvUserID)
			require.NoError(t, err)
			// Create Wallets
			recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
			require.NoError(t, err)

			asset := "USD"
			assetScale := 2
			if tc.args.IncomingAmount != nil {
				asset = tc.args.IncomingAmount.Asset
				assetScale = tc.args.IncomingAmount.AssetScale
			}
			if tc.ppAsset != "" {
				asset = tc.ppAsset
			}
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.PaymentPointer,
				WalletID:   recvWallet.ID,
				Alias:      "Alias",
				Asset:      asset,
				AssetScale: assetScale,
			})
			require.NoError(t, err)

			ip, err := ops.CreateIncomingPayment(ctx, b, tc.args)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.args.PaymentPointer, ip.PaymentPointer)
			assert.Equal(t, tc.args.ExternalRef, ip.ExternalRef)
			if tc.args.IncomingAmount != nil {
				assert.Equal(t, tc.args.IncomingAmount.Asset, ip.IncomingAmount.Asset)
				assert.Equal(t, tc.args.IncomingAmount.AssetScale, ip.IncomingAmount.AssetScale)
				assert.Equal(t, tc.args.IncomingAmount.Value, ip.IncomingAmount.Value)
			}
		})
	}
}

func TestCreateOutgoingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name      string
		quoteArgs openpayments.CreateQuoteArgs
		opArgs    openpayments.CreateOutgoingPaymentArgs
		err       error
	}{
		{
			name: "success",
			quoteArgs: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "http://fynbos.me/paysend",
				ReceivePaymentPointer: "http://fynbos.me/payrecv",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
			},
			opArgs: openpayments.CreateOutgoingPaymentArgs{
				Description: "Description",
				ExternalRef: "ExternalRef",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sendUserID := uuid.NewString()
			recvUserID := uuid.NewString()
			// Create Signups
			_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), sendUserID, uuid.NewString(), recvUserID)
			require.NoError(t, err)
			// Create Wallets
			sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
			require.NoError(t, err)
			recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.SendPaymentPointer,
				WalletID:   sendWallet.ID,
				Alias:      "Alias",
				Asset:      tc.quoteArgs.SendAmount.Asset,
				AssetScale: tc.quoteArgs.SendAmount.AssetScale,
			})
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.ReceivePaymentPointer,
				WalletID:   recvWallet.ID,
				Alias:      "Alias",
				Asset:      tc.quoteArgs.SendAmount.Asset,
				AssetScale: tc.quoteArgs.SendAmount.AssetScale,
			})
			require.NoError(t, err)

			q, err := ops.CreateQuote(ctx, b, tc.quoteArgs)
			require.NoError(t, err)

			tc.opArgs.QuoteID = q.ID
			opID, err := ops.CreateOutgoingPayment(ctx, b, tc.opArgs)
			require.NoError(t, err)

			op, err := ops.GetOutgoingPayment(ctx, b, opID)
			require.NoError(t, err)

			assert.Equal(t, opID, op.ID)
			assert.Equal(t, tc.quoteArgs.SendPaymentPointer, op.PaymentPointer)
			assert.Equal(t, tc.opArgs.Description, op.Description)
			assert.True(t, strings.HasPrefix(op.Receiver, tc.quoteArgs.ReceivePaymentPointer))
		})
	}
}
