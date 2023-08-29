package ops_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"
	identity_mock "gitlab.com/fynbos/backend/identities/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
	"gitlab.com/fynbos/backend/providers/tabapay"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	temporal_client "go.temporal.io/sdk/client"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Ic:  identity_mock.NewMockClient(ctrl),
		Wc:  wallets_mock.NewMockClient(ctrl),
		Lac: linkedaccounts_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: tabapay.ProviderName, State: linkedaccounts.Verified}, nil).AnyTimes()
	b.Lac.EXPECT().GetDefaultReceive(ctx, gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()
	b.Ic.EXPECT().GetByIdentifier(ctx, gomock.Any()).Return(&identities.Identity{WalletID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://fynbos.me/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	cases := []struct {
		name    string
		args    payments.CreateArgs
		actions []payments.RequiredActionType
		err     error
	}{
		{
			name: "success",
			args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeTwitter,
					Identifier: "@willy_wonka",
				},
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletURL,
					Identifier: "https://fynbos.me/charlie",
				},
				SenderAmount:    currency.FromFloat64(51, currency.USD),
				ReceiverAmount:  currency.FromFloat64(50, currency.USD),
				SenderAccount:   uuid.NewString(),
				ReceiverAccount: uuid.NewString(),
				Note:            "This is a NOTE!!!",
				IPAddress:       "193.9.4.6",
			},
			actions: []payments.RequiredActionType{payments.RequiredActionTypeThreeDS},
			err:     nil,
		},
		{
			name: "success_no_accounts",
			args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeTwitter,
					Identifier: "@willy_wonka",
				},
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletURL,
					Identifier: "https://fynbos.me/charlie",
				},
				SenderAmount:   currency.FromFloat64(51, currency.USD),
				ReceiverAmount: currency.FromFloat64(50, currency.USD),
			},
			actions: []payments.RequiredActionType{payments.RequiredActionTypeThreeDS, payments.RequiredActionTypeSenderAccount, payments.RequiredActionTypeIPAddress},
			err:     nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ops.Create(ctx, b, tc.args)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, payments.IdentityTypeWalletID, p.Sender.Type)
			assert.Equal(t, walletID, p.Sender.Identifier)
			assert.Equal(t, tc.args.Receiver.Type, p.Receiver.Type)
			assert.Equal(t, tc.args.Receiver.Identifier, p.Receiver.Identifier)
			assert.Equal(t, tc.args.SenderAccount, p.SenderAccount)
			assert.Equal(t, tc.args.ReceiverAccount, p.ReceiverAccount)
			assert.Equal(t, tc.args.SenderAmount.Format(), p.SenderAmount.Format())
			assert.Equal(t, tc.args.ReceiverAmount.Format(), p.ReceiverAmount.Format())
			assert.Equal(t, tc.args.Note, p.Note)
			assert.Equal(t, tc.args.IPAddress, p.IPAddress)
			assert.Len(t, p.RequiredActions, len(tc.actions))
			for _, ra := range tc.actions {
				assert.Contains(t, p.RequiredActions, ra)
			}
		})
	}
}

func TestSetState(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Ic:  identity_mock.NewMockClient(ctrl),
		Wc:  wallets_mock.NewMockClient(ctrl),
		Lac: linkedaccounts_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: tabapay.ProviderName, State: linkedaccounts.Verified}, nil).AnyTimes()
	b.Ic.EXPECT().GetByIdentifier(ctx, gomock.Any()).Return(&identities.Identity{WalletID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://fynbos.me/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://fynbos.me/charlie",
		},
		SenderAmount:    currency.FromFloat64(51, currency.USD),
		ReceiverAmount:  currency.FromFloat64(50, currency.USD),
		SenderAccount:   uuid.NewString(),
		ReceiverAccount: uuid.NewString(),
		IPAddress:       "193.9.4.6",
	})
	require.NoError(t, err)

	assert.ErrorIs(t, ops.SetState(ctx, b, p.ID, payments.StateCompleted), payments.ErrInvalidStateTransition)
	assert.NoError(t, ops.SetState(ctx, b, p.ID, payments.StateConfirmed))
}

func TestGetRequiredActions(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Wc:  wallets_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		RequiresOTP: true,
	})
	require.NoError(t, err)

	requiredActions, err := ops.GetRequiredActions(ctx, b, p.ID)
	require.NoError(t, err)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAccount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeOTP)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeThreeDS)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeIPAddress)
}

func TestConfirm(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Tp:  temporal_mock.NewMockClient(ctrl),
		Wc:  wallets_mock.NewMockClient(ctrl),
		Lac: linkedaccounts_mock.NewMockClient(ctrl),
		Txc: transactions_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	txID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: tabapay.ProviderName, State: linkedaccounts.Verified}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://fynbos.me/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		RequiresOTP: true,
		IPAddress:   "193.9.4.6",
	})
	require.NoError(t, err)
	paymentID := p.ID

	_, requiredActions, err := ops.Confirm(ctx, b, paymentID)
	require.Error(t, err)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAccount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeOTP)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeThreeDS)

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID: paymentID,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://fynbos.me/charlie",
		},
		SenderAmount:    currency.FromFloat64(51, currency.USD),
		ReceiverAmount:  currency.FromFloat64(50, currency.USD),
		SenderAccount:   uuid.NewString(),
		ReceiverAccount: uuid.NewString(),
		OTP:             "123",
		ThreeDSID:       "123",
	})
	require.NoError(t, err)
	assert.Equal(t, payments.StateCreated, p.State)

	b.Tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, wo temporal_client.StartWorkflowOptions, w interface{}, args ...interface{}) (temporal_client.WorkflowRun, error) {
			assert.Len(t, args, 1)
			assert.Equal(t, p.ID, args[0])
			return nil, nil
		},
	).Times(1)

	p, requiredActions, err = ops.Confirm(ctx, b, paymentID)
	require.NoError(t, err)
	assert.Empty(t, requiredActions)
	assert.Equal(t, payments.StateConfirmed, p.State)
	assert.Equal(t, txID, p.SendTransactionID)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		DBC: db.MigrateTestDB(t, ctx),
		Wc:  wallets_mock.NewMockClient(ctrl),
		Lac: linkedaccounts_mock.NewMockClient(ctrl),
	}
	walletID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: tabapay.ProviderName, State: linkedaccounts.Verified}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://fynbos.me/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()

	senderAccount := uuid.NewString()
	receiverAccount := uuid.NewString()
	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		SenderAmount:    currency.FromFloat64(51, currency.USD),
		ReceiverAmount:  currency.FromFloat64(50, currency.USD),
		SenderAccount:   senderAccount,
		ReceiverAccount: receiverAccount,
	})
	require.NoError(t, err)
	paymentID := p.ID

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID: paymentID,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://fynbos.me/charlie",
		},
	})
	require.NoError(t, err)
	assert.True(t, p.Receiver.IsEqual(payments.Identity{
		Type:       payments.IdentityTypeWalletURL,
		Identifier: "https://fynbos.me/charlie",
	}))
	assert.True(t, p.SenderAmount.IsEqual(currency.FromFloat64(51, currency.USD)))
	assert.True(t, p.ReceiverAmount.IsEqual(currency.FromFloat64(50, currency.USD)))
	assert.Equal(t, receiverAccount, p.ReceiverAccount)
	assert.Equal(t, senderAccount, p.SenderAccount)

	newSenderAccount := uuid.NewString()
	newReceiverAccount := uuid.NewString()
	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID:              paymentID,
		SenderAmount:    currency.FromFloat64(53, currency.USD),
		ReceiverAmount:  currency.FromFloat64(54, currency.USD),
		SenderAccount:   newSenderAccount,
		ReceiverAccount: newReceiverAccount,
	})
	require.NoError(t, err)

	assert.True(t, p.Receiver.IsEqual(payments.Identity{
		Type:       payments.IdentityTypeWalletURL,
		Identifier: "https://fynbos.me/charlie",
	}))
	assert.True(t, p.SenderAmount.IsEqual(currency.FromFloat64(53, currency.USD)))
	assert.True(t, p.ReceiverAmount.IsEqual(currency.FromFloat64(54, currency.USD)))
	assert.Equal(t, newReceiverAccount, p.ReceiverAccount)
	assert.Equal(t, newSenderAccount, p.SenderAccount)
}
