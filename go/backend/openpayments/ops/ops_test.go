package ops_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/mx"

	"gitlab.com/fynbos/backend/providers/gmt"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"gitlab.com/fynbos/backend/linkedaccounts"

	linked_account_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/golang/mock/gomock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
)

func TestCreatePaymentPointer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

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
			url:       "https://fynbos.me/abcd1",
			alias:     "Alias",
			assetCode: "USD",
			scale:     2,
			err:       nil,
		},
		{
			name:      "invalid_url",
			url:       "httpssss://fynbos.me/creature",
			alias:     "Alias",
			assetCode: "USD",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
		},
		{
			name:      "invalid_asset",
			url:       "https://fynbos.me/abcd2",
			alias:     "Alias",
			assetCode: "FUzzY",
			scale:     2,
			err:       openpayments.ErrInvalidArgument,
		},
		{
			name:      "duplicate",
			url:       "https://fynbos.me/abcd3",
			alias:     "Alias",
			assetCode: "ZAR",
			scale:     2,
			duplicate: true,
			err:       openpayments.ErrPaymentPointerExists,
		},
		{
			name:      "regex_first_4_not_alpha",
			url:       "https://fynbos.me/1234PayMe",
			alias:     "Alias",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your first 3 characters must be letters",
		},
		{
			name:      "regex_contains_slash",
			url:       "https://fynbos.me/PayMe/1234",
			alias:     "Alias",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:      "regex_too_short",
			url:       "https://fynbos.me/Pa",
			alias:     "Alias",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer must be longer than 3 characters",
		},
		{
			name:      "regex_too_long",
			url:       "https://fynbos.me/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			alias:     "Alias",
			assetCode: "ZAR",
			scale:     2,
			err:       openpayments.ErrInvalidPointerPath,
			errMsg:    "Your payment pointer must be shorter than 30 characters",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			walletID := uuid.NewString()

			err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.url,
				WalletID:   walletID,
				Alias:      tc.alias,
				Asset:      currency.ParseCurrency(tc.assetCode),
				AssetScale: tc.scale,
			})
			if tc.duplicate {
				require.NoError(t, err)
				err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
					URL:        tc.url,
					WalletID:   walletID,
					Alias:      tc.alias,
					Asset:      currency.ParseCurrency(tc.assetCode),
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
			assert.Equal(t, walletID, pp.WalletID)
			assert.Equal(t, tc.url, pp.URL)
			assert.Equal(t, tc.assetCode, pp.Asset.String())
			assert.Equal(t, tc.scale, pp.AssetScale)
		})
	}
}

func TestGetWalletPaymentPointer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

	walletID := uuid.NewString()

	err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/abcd1",
		WalletID:   walletID,
		Alias:      "Alias",
		Asset:      currency.ParseCurrency("USD"),
		AssetScale: 2,
	})
	require.NoError(t, err)

	pp, err := ops.GetWalletPaymentPointer(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, "https://fynbos.me/abcd1", pp.URL)
	assert.Equal(t, "Alias", pp.Alias)
	assert.Equal(t, 2, pp.AssetScale)
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
			name:            "no_trailing_slash",
			url:             "https://fynbos.me/asdf",
			expectedPointer: "https://fynbos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name:            "trailing_slash",
			url:             "https://fynbos.me/asdf/",
			expectedPointer: "https://fynbos.me/asdf",
			expectedSuffix:  "",
			err:             nil,
		},
		{
			name: "invalid_url",
			url:  "ladida-blah",
			err:  openpayments.ErrInvalidPointerURL,
		},
		{
			name:            "extracts_suffix_jwks.json",
			url:             "https://fynbos.me/asdf/jwks.json",
			expectedPointer: "https://fynbos.me/asdf",
			expectedSuffix:  "jwks.json",
			err:             nil,
		},
		{
			name:            "extracts_suffix_incoming",
			url:             "https://fynbos.me/asdf/incoming",
			expectedPointer: "https://fynbos.me/asdf",
			expectedSuffix:  "incoming",
			err:             nil,
		},
		{
			name:            "double_suffix",
			url:             "https://fynbos.me/asdf/bbsdf",
			expectedPointer: "https://fynbos.me/asdf/bbsdf",
			expectedSuffix:  "",
			err:             nil,
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
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

	walletID := uuid.NewString()

	err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/payp1",
		WalletID:   walletID,
		Alias:      "Alias1",
		Asset:      "USD",
		AssetScale: 2,
	})
	require.NoError(t, err)
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/payp2",
		WalletID:   walletID,
		Alias:      "Alias2",
		Asset:      "ZAR",
		AssetScale: 2,
	})
	require.NoError(t, err)

	// List and validate
	pp, err := ops.ListWalletPaymentPointers(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, pp, 2)
	for _, p := range pp {
		if p.URL == "https://fynbos.me/payp1" {
			assert.Equal(t, "Alias1", p.Alias)
			assert.Equal(t, "USD", p.Asset.String())
			assert.Equal(t, 2, p.AssetScale)
		} else {
			assert.Equal(t, "Alias2", p.Alias)
			assert.Equal(t, "ZAR", p.Asset.String())
			assert.Equal(t, 2, p.AssetScale)
		}
	}
}

func TestValidatePaymentPointer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

	cases := []struct {
		name   string
		url    string
		err    error
		errMsg string
		exists bool
	}{
		{
			name: "success",
			url:  "https://fynbos.me/abcd1",
			err:  nil,
		},
		{
			name: "reserved_word",
			url:  "fynbos.me/incoming",
			err:  openpayments.ErrInvalidPointerURL,
		},
		{
			name:   "regex_first_4_not_alpha",
			url:    "https://fynbos.me/1234PayMe",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your first 3 characters must be letters",
		},
		{
			name:   "regex_contains_slash",
			url:    "https://fynbos.me/PayMe/1234",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "regex_too_short",
			url:    "https://fynbos.me/Pa",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer must be longer than 3 characters",
		},
		{
			name:   "regex_too_long",
			url:    "https://fynbos.me/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer must be shorter than 30 characters",
		},
		{
			name:   "contains_hash",
			url:    "https://fynbos.me/asdfn#asdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_space",
			url:    "https://fynbos.me/asdfn%20asdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_<",
			url:    "https://fynbos.me/asdfn%3Casdf",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "trailing_slash",
			url:    "https://fynbos.me/asdfn/",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "query_params",
			url:    "https://fynbos.me/asdfn?asdf3e=34334",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_escapes",
			url:    "https://fynbos.me/asdfn%2",
			err:    openpayments.ErrInvalidPointerPath,
			errMsg: "Your payment pointer can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_dots",
			url:    "https://fynbos.me/abcdef..sasd",
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
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

	walletID := uuid.NewString()

	exists, err := ops.PaymentPointerExists(ctx, b, "https://fynbos.me/ValidPaymentPointer")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create payment pointer
	err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
		URL:        "https://fynbos.me/ValidPaymentPointer",
		WalletID:   walletID,
		Asset:      "ZAR",
		Alias:      "test",
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
		WalletID:   walletID,
		Asset:      "ZAR",
		Alias:      "test",
		AssetScale: 2,
	})
	require.ErrorIs(t, err, openpayments.ErrPaymentPointerExists)
}

func TestStandardisePaymentPointer(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "http",
			url:      "https://fynbos.me/asdf",
			expected: "https://fynbos.me/asdf",
		},
		{
			name:     "https",
			url:      "https://fynbos.me/asdf",
			expected: "https://fynbos.me/asdf",
		},
		{
			name:     "dollar",
			url:      "$fynbos.me/asdf",
			expected: "https://fynbos.me/asdf",
		},
		{
			name:     "noprefix",
			url:      "fynbos.me/asdf",
			expected: "https://fynbos.me/asdf",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			standard := ops.StandardisePaymentPointer(tc.url)
			assert.Equal(t, tc.expected, standard)
		})
	}
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
			url:               "https://fynbos.me/asdf",
			expectedFormatted: "fynbos.me/asdf",
			err:               nil,
		},
		{
			name:              "https",
			url:               "https://fynbos.me/asdf",
			expectedFormatted: "fynbos.me/asdf",
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
	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	laClient := linked_account_mock.NewMockClient(ctrl)
	txClient := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	txClient.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	txClient.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	b := ops.NewTestBackends(t, db, laClient, txClient)

	cases := []struct {
		name      string
		args      openpayments.CreateQuoteArgs
		recvAsset string
		balance   uint64
		err       error
	}{
		{
			name: "success",
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend",
				ReceivePaymentPointer: "https://fynbos.me/payrecv",
				Description:           "IncomingPayment",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				CreatedBy: uuid.NewString(),
			},
		},
		{
			name:      "different assets",
			recvAsset: "ZAR",
			err:       openpayments.ErrInvalidArgument,
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend1",
				ReceivePaymentPointer: "https://fynbos.me/payrecv2",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
			},
		},
		{
			name: "expiry in the past",
			err:  openpayments.ErrInvalidArgument,
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend3",
				ReceivePaymentPointer: "https://fynbos.me/payrecv4",
				ExpiresAt:             time.Now().Add(time.Hour * -1),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
			},
		},
		{
			name: "success send from wallet",
			args: openpayments.CreateQuoteArgs{
				SendPaymentPointer:    "https://fynbos.me/paysend9",
				ReceivePaymentPointer: "https://fynbos.me/payrecv10",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				LinkedAccID: "4153f92e-158a-46dc-b298-4b71635c2093",
			},
			balance: 10000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sendWalletID := uuid.NewString()
			recvWalletID := uuid.NewString()

			err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.SendPaymentPointer,
				WalletID:   sendWalletID,
				Alias:      "Alias",
				Asset:      tc.args.SendAmount.Currency,
				AssetScale: tc.args.SendAmount.Scale,
			})
			require.NoError(t, err)
			recvAsset := tc.args.SendAmount.Currency
			if tc.recvAsset != "" {
				recvAsset = currency.ParseCurrency(tc.recvAsset)
			}
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.ReceivePaymentPointer,
				WalletID:   recvWalletID,
				Alias:      "Alias",
				Asset:      recvAsset,
				AssetScale: tc.args.SendAmount.Scale,
			})
			require.NoError(t, err)

			if tc.args.LinkedAccID != "" {
				providerID := uuid.NewString()
				_, err = db.ExecContext(ctx, "insert into linked_accounts (id, wallet_id, name, mask, provider, provider_id, type, can_send, can_receive) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
					tc.args.LinkedAccID, sendWalletID, "testing", "mask", gmt.ProviderName, providerID, gmt.TypeBankAccount, true, true)
				require.NoError(t, err)
				laClient.EXPECT().Get(ctx, tc.args.LinkedAccID).Return(&linkedaccounts.LinkedAccount{
					ID:         tc.args.LinkedAccID,
					WalletID:   sendWalletID,
					ProviderID: providerID,
					Type:       gmt.TypeBankAccount,
				}, nil)

				laClient.EXPECT().ListByWalletId(ctx, sendWallet.ID).Return([]linkedaccounts.LinkedAccount{
					{
						ID:         tc.args.LinkedAccID,
						WalletID:   sendWallet.ID,
						Name:       "NoName",
						Nickname:   "NoName",
						Mask:       "1234",
						Provider:   mx.ProviderName,
						ProviderID: providerID,
						Type:       "card",
						CanSend:    true,
						CanReceive: true,
						State:      linkedaccounts.Verified,
					},
				}, nil).AnyTimes()
			} else {
				laClient.EXPECT().ListByWalletId(ctx, sendWallet.ID).Return([]linkedaccounts.LinkedAccount{
					{
						ID:         uuid.NewString(),
						WalletID:   sendWallet.ID,
						Name:       "NoName",
						Nickname:   "NoName",
						Mask:       "1234",
						Provider:   mx.ProviderName,
						ProviderID: uuid.NewString(),
						Type:       "card",
						CanSend:    true,
						CanReceive: true,
						State:      linkedaccounts.Verified,
					},
				}, nil).AnyTimes()
			}

			laClient.EXPECT().ListByWalletId(ctx, recvWallet.ID).Return([]linkedaccounts.LinkedAccount{
				{
					ID:         uuid.NewString(),
					WalletID:   recvWallet.ID,
					Name:       "NoName",
					Nickname:   "NoName",
					Mask:       "1234",
					Provider:   mx.ProviderName,
					ProviderID: uuid.NewString(),
					Type:       "card",
					CanSend:    true,
					CanReceive: true,
					State:      linkedaccounts.Verified,
				},
			}, nil).AnyTimes()

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
			assert.Equal(t, tc.args.CreatedBy, q.CreatedBy)
			assert.Equal(t, tc.args.SendAmount.Value, q.ReceiveAmount.Value)
			assert.Equal(t, tc.args.SendAmount.Currency, q.ReceiveAmount.Currency)
			assert.Equal(t, tc.args.SendAmount.Scale, q.ReceiveAmount.Scale)
			assert.Equal(t, tc.args.SendAmount.Value, q.SendAmount.Value)
			assert.Equal(t, tc.args.SendAmount.Currency, q.SendAmount.Currency)
			assert.Equal(t, tc.args.SendAmount.Scale, q.SendAmount.Scale)
		})
	}
}

func TestValidateCanSend(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	laClient := linked_account_mock.NewMockClient(ctrl)
	txClient := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	txClient.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	b := ops.NewTestBackends(t, db, laClient, txClient)

	cases := []struct {
		name      string
		sendPP    string
		recvPP    string
		unauthed  bool
		hasWallet bool
		hasPP     bool
		expected  bool
		err       error
	}{
		{
			name:      "can_send",
			sendPP:    "https://fynbos.me/paysend",
			recvPP:    "https://fynbos.me/payrecv",
			hasWallet: true,
			hasPP:     true,
			expected:  true,
		},
		{
			name:      "same_paymentpointer",
			sendPP:    "https://fynbos.me/samesame",
			recvPP:    "https://fynbos.me/samesame",
			hasWallet: false,
			hasPP:     true,
			expected:  false,
		},
		{
			name:      "no_such_pointer",
			sendPP:    "https://fynbos.me/senderpoint",
			recvPP:    "https://fynbos.me/notreal",
			hasWallet: false,
			hasPP:     false,
			expected:  false,
		},
		{
			name:      "dollar_sign_pointer",
			sendPP:    "$fynbos.me/tothedollar",
			recvPP:    "$fynbos.me/accepteddollar",
			hasWallet: true,
			hasPP:     true,
			expected:  true,
		},
		{
			name:      "un_authed_can_send",
			sendPP:    "$fynbos.me/notloggedin",
			recvPP:    "$fynbos.me/thisisreal",
			hasWallet: true,
			unauthed:  true,
			hasPP:     true,
			expected:  true,
		},
		{
			name:      "un_authed_no_pp",
			sendPP:    "$fynbos.me/notloggedinagainreallly",
			recvPP:    "$fynbos.me/stillnotreal",
			hasWallet: false,
			unauthed:  true,
			hasPP:     false,
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sendWalletID := uuid.NewString()
			recvWalletID := uuid.NewString()

			err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.sendPP,
				WalletID:   sendWalletID,
				Alias:      "Alias",
				Asset:      currency.USD,
				AssetScale: 2,
			})
			require.NoError(t, err)
			if tc.sendPP != tc.recvPP && tc.hasPP {
				err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
					URL:        tc.recvPP,
					WalletID:   recvWalletID,
					Alias:      "Alias",
					Asset:      currency.USD,
					AssetScale: 2,
				})
				require.NoError(t, err)
			}

			laClient.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
				{
					ID:         uuid.NewString(),
					Provider:   gmt.ProviderName,
					Type:       gmt.TypeBankAccount,
					CanSend:    true,
					CanReceive: true,
					State:      linkedaccounts.Verified,
				},
			}, nil).AnyTimes()

			walletID := sendWalletID
			if tc.unauthed {
				walletID = ""
			}
			canSend, err := ops.ValidateCanSend(ctx, b, walletID, tc.recvPP)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, canSend)
		})
	}
}
