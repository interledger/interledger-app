package ops_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/providers/mx"

	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

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
			err:  wallets.ErrInvalidAddress,
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

func TestPaymentPointerExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().GetFromAddress(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	b := ops.NewTestBackends(t, db, nil, nil, func(b *ops.TestBackends) {
		b.Wc = wc
	})

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
			err:  wallets.ErrInvalidAddress,
		},
		{
			name:   "regex_first_4_not_alpha",
			url:    "https://fynbos.me/1234PayMe",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your first 3 characters must be letters",
		},
		{
			name:   "regex_contains_slash",
			url:    "https://fynbos.me/PayMe/1234",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "regex_too_short",
			url:    "https://fynbos.me/Pa",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address must be longer than 3 characters",
		},
		{
			name:   "regex_too_long",
			url:    "https://fynbos.me/asdfnwelkjnasfdgoiaertaqri0943lnsfgas094905",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address must be shorter than 30 characters",
		},
		{
			name:   "contains_hash",
			url:    "https://fynbos.me/asdfn#asdf",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_space",
			url:    "https://fynbos.me/asdfn%20asdf",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "contains_url_escape_<",
			url:    "https://fynbos.me/asdfn%3Casdf",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "query_params",
			url:    "https://fynbos.me/asdfn?asdf3e=34334",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_escapes",
			url:    "https://fynbos.me/asdfn%2",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
		{
			name:   "invalid_dots",
			url:    "https://fynbos.me/abcdef..sasd",
			err:    wallets.ErrInvalidAddress,
			errMsg: "Your wallet address can only contain letters, numbers and '_'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			exists, err := ops.PaymentPointerExists(ctx, b, tc.url)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				fmt.Println("XXXXX", err)
				assert.True(t, strings.HasSuffix(err.Error(), tc.errMsg))
				assert.True(t, strings.HasPrefix(err.Error(), tc.err.Error()))
				return
			}

			require.Equal(t, tc.exists, exists)
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
	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().AddAddress(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	b := ops.NewTestBackends(t, db, laClient, txClient, func(b *ops.TestBackends) {
		b.Wc = wc
	})

	cases := []struct {
		name    string
		args    openpayments.CreateQuoteArgs
		balance uint64
		err     error
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

			sa, err := wallets.ParseAddress(tc.args.SendPaymentPointer)
			require.NoError(t, err)
			wc.EXPECT().GetFromAddress(gomock.Any(), tc.args.SendPaymentPointer).Return(&wallets.Wallet{
				ID:        sendWalletID,
				Addresses: []wallets.Address{sa},
			}, nil).AnyTimes()

			ra, err := wallets.ParseAddress(tc.args.ReceivePaymentPointer)
			require.NoError(t, err)
			wc.EXPECT().GetFromAddress(gomock.Any(), tc.args.ReceivePaymentPointer).Return(&wallets.Wallet{
				ID:        recvWalletID,
				Addresses: []wallets.Address{ra},
			}, nil).AnyTimes()

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

				laClient.EXPECT().ListByWalletId(ctx, sendWalletID).Return([]linkedaccounts.LinkedAccount{
					{
						ID:         tc.args.LinkedAccID,
						WalletID:   sendWalletID,
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
				laClient.EXPECT().ListByWalletId(ctx, sendWalletID).Return([]linkedaccounts.LinkedAccount{
					{
						ID:         uuid.NewString(),
						WalletID:   sendWalletID,
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

			laClient.EXPECT().ListByWalletId(ctx, recvWalletID).Return([]linkedaccounts.LinkedAccount{
				{
					ID:         uuid.NewString(),
					WalletID:   recvWalletID,
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
	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().AddAddress(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	b := ops.NewTestBackends(t, db, laClient, txClient, func(b *ops.TestBackends) {
		b.Wc = wc
	})

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

			sa, err := wallets.ParseAddress(tc.sendPP)
			require.NoError(t, err)
			wc.EXPECT().Get(gomock.Any(), sendWalletID).Return(&wallets.Wallet{
				ID:        sendWalletID,
				Addresses: []wallets.Address{sa},
			}, nil).AnyTimes()

			ra, err := wallets.ParseAddress(tc.recvPP)
			require.NoError(t, err)
			if tc.hasPP {
				wc.EXPECT().GetFromAddress(gomock.Any(), tc.recvPP).Return(&wallets.Wallet{
					ID:        recvWalletID,
					Addresses: []wallets.Address{ra},
				}, nil).AnyTimes()
			} else {
				wc.EXPECT().GetFromAddress(gomock.Any(), tc.recvPP).Return(nil, wallets.ErrNoWalletFound).AnyTimes()
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
