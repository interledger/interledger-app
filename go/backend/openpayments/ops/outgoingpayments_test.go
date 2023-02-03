package ops_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/currency"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock "gitlab.com/fynbos/backend/providers/machnet/client/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
)

func TestCreateOutgoingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	db := db.MigrateTestDB(t, ctx)

	mc := machnet_mock.NewMockClient(ctrl)
	mc.EXPECT().GetUserByWalletID(gomock.Any(), gomock.Any()).Return(&machnet.User{KYCStatus: machnet.KYCStatusVerified}, nil).AnyTimes()

	txClient := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	txClient.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	b := ops.NewTestBackends(t, db, nil, mc, txClient)

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
				SendPaymentPointer:    "https://fynbos.me/paysend",
				ReceivePaymentPointer: "https://fynbos.me/payrecv",
				ExpiresAt:             time.Now().Add(time.Hour),
				SendAmount: currency.Amount{
					Value:    100,
					Currency: "USD",
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

			op, err := ops.GetOutgoingPayment(ctx, b, opID)
			require.NoError(t, err)

			assert.Equal(t, opID, op.ID)
			assert.Equal(t, tc.quoteArgs.SendPaymentPointer, op.PaymentPointer)
			assert.Equal(t, tc.opArgs.Description, op.Description)
			assert.True(t, strings.HasPrefix(op.Receiver, tc.quoteArgs.ReceivePaymentPointer))
		})
	}
}
