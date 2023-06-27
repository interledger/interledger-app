package ops_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/currency"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
)

func TestCreateOutgoingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	db := db.MigrateTestDB(t, ctx)

	txClient := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	txClient.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	txClient.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	laClient := linkedaccounts_mock.NewMockClient(ctrl)
	laClient.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{
			ID:         uuid.NewString(),
			WalletID:   uuid.NewString(),
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
	b := ops.NewTestBackends(t, db, laClient, txClient)

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
				CreatedBy: uuid.NewString(),
			},
			opArgs: openpayments.CreateOutgoingPaymentArgs{
				Description: "Description",
				CreatedBy:   uuid.NewString(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sendWalletID := uuid.NewString()
			recvWalletID := uuid.NewString()

			err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.SendPaymentPointer,
				WalletID:   sendWalletID,
				Alias:      "Alias",
				Asset:      tc.quoteArgs.SendAmount.Currency,
				AssetScale: tc.quoteArgs.SendAmount.Scale,
			})
			require.NoError(t, err)

			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.quoteArgs.ReceivePaymentPointer,
				WalletID:   recvWalletID,
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
			assert.Equal(t, tc.opArgs.CreatedBy, op.CreatedBy)
			assert.True(t, strings.HasPrefix(op.Receiver, tc.quoteArgs.ReceivePaymentPointer))
		})
	}
}
