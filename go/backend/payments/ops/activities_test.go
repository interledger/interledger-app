package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/providers/pti"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"

	"gitlab.com/fynbos/backend/linkedaccounts"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/identities"
	id_mock "gitlab.com/fynbos/backend/identities/client/mock"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
	"go.temporal.io/sdk/temporal"
)

func TestSetPaymentState(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Em:  email_mock.NewMockClient(ctrl),
		Ic:  id_mock.NewMockClient(ctrl),
		Wc:  wallets_mock.NewMockClient(ctrl),
		Lac: linkedaccounts_mock.NewMockClient(ctrl),
		Txc: transactions_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: pti.ProviderName, State: linkedaccounts.Verified, ReceiveCurrency: currency.USD, SendCurrency: currency.USD}, nil).AnyTimes()
	b.Ic.EXPECT().GetByIdentifier(ctx, gomock.Any()).Return(&identities.Identity{
		WalletID: walletID,
	}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes() // No OTP
	b.Lac.EXPECT().GetDefaultReceive(ctx, gomock.Any(), gomock.Any()).Return(&linkedaccounts.LinkedAccount{ID: uuid.NewString(), WalletID: walletID}, nil).AnyTimes()
	a := ops.NewActivity(b)

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
		},
		SenderAmount:    currency.FromFloat64(51, currency.USD),
		ReceiverAmount:  currency.FromFloat64(50, currency.USD),
		SenderAccount:   uuid.NewString(),
		ReceiverAccount: uuid.NewString(),
		IPAddress:       "193.9.4.6",
	})
	require.NoError(t, err)

	t.Run("Sends sent and received email when payment is complete", func(st *testing.T) {
		_, err := b.DBC.ExecContext(ctx, "UPDATE payments set state=$1 WHERE id=$2;", payments.StateProcessing, p.ID)
		require.NoError(st, err)

		b.Em.EXPECT().SendPaymentSentEmailV2(ctx, gomock.Any(), gomock.Any()).Do(
			func(ctx context.Context, id string, payment *payments.Payment) {
				assert.Equal(st, walletID, id, "walletID incorrect")
				assert.Equal(st, p.ID, payment.ID, "paymentID incorrect")
			},
		).Times(1)

		b.Em.EXPECT().SendPaymentReceivedEmailV2(ctx, gomock.Any(), gomock.Any()).Do(
			func(ctx context.Context, id string, payment *payments.Payment) {
				assert.Equal(st, walletID, id, "walletID incorrect")
				assert.Equal(st, p.ID, payment.ID, "paymentID incorrect")
			},
		).Times(1)

		err = a.SetPaymentStateComplete(ctx, p.ID)
		assert.NoError(st, err)

		err = a.SetPaymentStateComplete(ctx, p.ID)
		var applicationError *temporal.ApplicationError
		require.ErrorAs(t, err, &applicationError)
		assert.Equal(st, "ErrInvalidStateTransition", applicationError.Type())
	})

	t.Run("sends failed payment email when payment has failed", func(st *testing.T) {
		_, err := b.DBC.ExecContext(ctx, "UPDATE payments set state=$1 WHERE id=$2;", payments.StateProcessing, p.ID)
		require.NoError(st, err)

		b.Em.EXPECT().SendPaymentFailedEmail(ctx, gomock.Any()).Do(
			func(ctx context.Context, id string) {
				assert.Equal(st, walletID, id, "walletID incorrect")
			},
		).Times(1)

		err = a.SetPaymentStateFailed(ctx, p.ID)
		assert.NoError(st, err)

		err = a.SetPaymentStateFailed(ctx, p.ID)
		var applicationError *temporal.ApplicationError
		require.ErrorAs(t, err, &applicationError)
		assert.Equal(st, "ErrInvalidStateTransition", applicationError.Type())
	})

	t.Run("sets state to processing", func(st *testing.T) {
		_, err := b.DBC.ExecContext(ctx, "UPDATE payments set state=$1 WHERE id=$2;", payments.StateConfirmed, p.ID)
		require.NoError(st, err)

		err = a.SetPaymentStateProcessing(ctx, p.ID)
		assert.NoError(st, err)

		err = a.SetPaymentStateProcessing(ctx, p.ID)
		var applicationError *temporal.ApplicationError
		require.ErrorAs(t, err, &applicationError)
		assert.Equal(st, "ErrInvalidStateTransition", applicationError.Type())
	})
}
