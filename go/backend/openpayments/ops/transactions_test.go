package ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
)

func TestListTransactions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, dbc, nil, nil)

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
			_, err := dbc.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), sendUserID, uuid.NewString(), recvUserID)
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
			_, err = ops.CreateOutgoingPayment(ctx, b, tc.opArgs)
			require.NoError(t, err)

			pendingSender, err := ops.ListPendingTransactions(ctx, b, sendWallet.ID, db.Pagination{})
			require.NoError(t, err)
			assert.Len(t, pendingSender, 1)

			assert.Equal(t, pendingSender[0].Source, tc.quoteArgs.SendPaymentPointer)
			assert.Equal(t, pendingSender[0].Destination, tc.quoteArgs.ReceivePaymentPointer)
			assert.Equal(t, pendingSender[0].Type, openpayments.TransactionTypeOutgoingPayment)
			assert.Equal(t, pendingSender[0].Amount.Value, tc.quoteArgs.SendAmount.Value)
			assert.Equal(t, pendingSender[0].Amount.AssetScale, tc.quoteArgs.SendAmount.AssetScale)
			assert.Equal(t, pendingSender[0].Amount.Asset, tc.quoteArgs.SendAmount.Asset)

			err = ops.CompleteOutgoingPayment(ctx, b, openpayments.CompleteOutgoingPaymentArgs{
				ID:         pendingSender[0].ID,
				SentAmount: tc.quoteArgs.SendAmount,
			})
			require.NoError(t, err)

			pendingSender, err = ops.ListPendingTransactions(ctx, b, sendWallet.ID, db.Pagination{})
			require.NoError(t, err)
			assert.Empty(t, pendingSender)

			completeSender, err := ops.ListTransactions(ctx, b, sendWallet.ID, db.Pagination{})
			require.NoError(t, err)
			assert.Len(t, completeSender, 1)

			assert.Equal(t, completeSender[0].Source, tc.quoteArgs.SendPaymentPointer)
			assert.Equal(t, completeSender[0].Destination, tc.quoteArgs.ReceivePaymentPointer)
			assert.Equal(t, completeSender[0].Type, openpayments.TransactionTypeOutgoingPayment)
			assert.Equal(t, completeSender[0].Note, tc.opArgs.Description)
			assert.Equal(t, completeSender[0].Amount.Value, tc.quoteArgs.SendAmount.Value)
			assert.Equal(t, completeSender[0].Amount.AssetScale, tc.quoteArgs.SendAmount.AssetScale)
			assert.Equal(t, completeSender[0].Amount.Asset, tc.quoteArgs.SendAmount.Asset)

			completeRecv, err := ops.ListTransactions(ctx, b, recvWallet.ID, db.Pagination{})
			require.NoError(t, err)
			assert.Len(t, completeRecv, 1)

			assert.Equal(t, completeRecv[0].Source, tc.quoteArgs.SendPaymentPointer)
			assert.Equal(t, completeRecv[0].Destination, tc.quoteArgs.ReceivePaymentPointer)
			assert.Equal(t, completeRecv[0].Type, openpayments.TransactionTypeIncomingPayment)
			assert.Equal(t, completeRecv[0].Note, tc.opArgs.Description)
			assert.Equal(t, completeRecv[0].Amount.Value, tc.quoteArgs.SendAmount.Value)
			assert.Equal(t, completeRecv[0].Amount.AssetScale, tc.quoteArgs.SendAmount.AssetScale)
			assert.Equal(t, completeRecv[0].Amount.Asset, tc.quoteArgs.SendAmount.Asset)
		})
	}
}
