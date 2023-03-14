package workflows

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/currency"

	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_account_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	users_client "gitlab.com/fynbos/backend/user/client"
	"go.temporal.io/sdk/testsuite"
)

func TestActivity_GetProviderArgs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)

	txClient := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	txClient.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	la_mock := linked_account_mock.NewMockClient(ctrl)

	b := NewTestBackends(t, db, nil, la_mock, txClient, nil, nil)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := Activity{b: b}

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	env.RegisterActivity(a.GetProviderArgs)

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
				SendAmount: currency.Amount{
					Value:    100,
					Currency: currency.USD,
					Scale:    2,
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
			// Create Wallets
			sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
			require.NoError(t, err)
			recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.SendPaymentPointer,
				WalletID:   sendWallet.ID,
				Alias:      "Alias",
				Asset:      tc.quoteArgs.SendAmount.Currency,
				AssetScale: tc.quoteArgs.SendAmount.Scale,
			})
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.ReceivePaymentPointer,
				WalletID:   recvWallet.ID,
				Alias:      "Alias",
				Asset:      tc.quoteArgs.SendAmount.Currency,
				AssetScale: tc.quoteArgs.SendAmount.Scale,
			})
			require.NoError(t, err)

			q, err := ops.CreateQuote(ctx, b, tc.quoteArgs)
			require.NoError(t, err)

			tc.opArgs.QuoteID = q.ID
			opID, _, err := ops.CreateOutgoingPayment(ctx, b, tc.opArgs)
			require.NoError(t, err)

			la_mock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
				{
					ID:       uuid.NewString(),
					Provider: machnet.ProviderName,
					Type:     machnet.TypeSendCard,
				}, {
					ID:       uuid.NewString(),
					Provider: machnet.ProviderName,
					Type:     machnet.TypeWallet,
				},
			}, nil).Times(2)

			argsEnc, err := env.ExecuteActivity(a.GetProviderArgs, opID)
			require.NoError(t, err)

			var args machnet.CreateTransactionArgs
			err = argsEnc.Get(&args)
			require.NoError(t, err)

			assert.Equal(t, 1.00, args.Amount.Float64())
			assert.Equal(t, "USD", args.Amount.Currency.String())
		})
	}
}
