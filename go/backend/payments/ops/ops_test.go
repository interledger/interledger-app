package ops_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	pti_mock "github.com/interledger/interledger-app/go/backend/providers/pti/client/mock"

	"github.com/interledger/interledger-app/go/backend/providers/pti"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/features"
	features_mock "github.com/interledger/interledger-app/go/backend/features/client/mock"
	"github.com/interledger/interledger-app/go/backend/identities"
	identity_mock "github.com/interledger/interledger-app/go/backend/identities/client/mock"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linkedaccounts_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/payments/ops"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	xago_mock "github.com/interledger/interledger-app/go/backend/providers/xago/client/mock"
	xago_external "github.com/interledger/interledger-app/go/backend/providers/xago/external"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	temporal_mock "github.com/interledger/interledger-app/go/backend/temporal/mock"
	transactions_mock "github.com/interledger/interledger-app/go/backend/transactions/client/mock"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	temporal_client "go.temporal.io/sdk/client"
)

func setupTest(t *testing.T) (context.Context, *ops.TestBackends) {

	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx, "")

	b := &ops.TestBackends{
		DBC: dbc,
		Ic:  identity_mock.NewMockClient(gomock.NewController(t)),
		Wc:  wallets_mock.NewMockClient(gomock.NewController(t)),
		Lac: linkedaccounts_mock.NewMockClient(gomock.NewController(t)),
		Txc: transactions_mock.NewMockClient(gomock.NewController(t)),
		Pti: pti_mock.NewMockClient(gomock.NewController(t)),
	}

	return ctx, b
}

func TestCreate(t *testing.T) {
	ctx, b := setupTest(t)

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	walletID := uuid.NewString()
	b.Lac.EXPECT().ListBalances(ctx, gomock.Any()).Return([]linkedaccounts.LinkedAccount{}, nil).AnyTimes()
	b.Pti.EXPECT().GetBalance(ctx, gomock.Any()).Return(&pti.Balance{Available: currency.FromFloat64(1000, currency.USD), Total: currency.FromFloat64(1000, currency.USD)}, nil).AnyTimes()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Type: pti.AccTypeBalance, Provider: pti.ProviderName, State: linkedaccounts.Verified, WalletID: walletID, SendCurrency: currency.USD, ReceiveCurrency: currency.USD}, nil).AnyTimes()
	b.Lac.EXPECT().GetDefaultReceive(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()
	b.Ic.EXPECT().GetByIdentifier(ctx, gomock.Any()).Return(&identities.Identity{WalletID: walletID, Platform: identities.PlatformTwitter}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes() // Require OTP
	b.Txc.EXPECT().CountSendTransactions(gomock.Any(), gomock.Any()).Return(60, nil).AnyTimes()             // for payment protection calcs
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
					Identifier: "https://ilp.link/charlie",
				},
				SenderAmount:    currency.FromFloat64(51, currency.USD),
				ReceiverAmount:  currency.FromFloat64(51, currency.USD),
				SenderAccount:   uuid.NewString(),
				ReceiverAccount: uuid.NewString(),
				Note:            "This is a NOTE!!!",
				IPAddress:       "193.9.4.6",
			},
			// actions: []payments.RequiredActionType{payments.RequiredActionTypeOTP},
			err: nil,
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
					Identifier: "https://ilp.link/charlie",
				},
				SenderAmount:   currency.FromFloat64(51, currency.USD),
				ReceiverAmount: currency.FromFloat64(51, currency.USD),
			},
			actions: []payments.RequiredActionType{payments.RequiredActionTypeSenderAccount, payments.RequiredActionTypeIPAddress},
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
			assert.Equal(t, tc.args.ReceiverAmount.Format(), p.ReceiverAmount.Format())
			assert.Equal(t, tc.args.Note, p.Note)
			assert.Equal(t, tc.args.IPAddress, p.IPAddress)
			assert.Len(t, p.RequiredActions, len(tc.actions))
			assert.Regexp(t, "^([b-z0-9]{12})$", p.PublicID)
			for _, ra := range tc.actions {
				assert.Contains(t, p.RequiredActions, ra)
			}

			assert.Equal(t, tc.args.SenderAmount.Format(), p.SenderAmount.Format())
		})
	}
}

func TestSetState(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	walletID := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: pti.ProviderName, State: linkedaccounts.Verified, ReceiveCurrency: currency.USD, SendCurrency: currency.USD}, nil).AnyTimes()
	b.Ic.EXPECT().GetByIdentifier(ctx, gomock.Any()).Return(&identities.Identity{WalletID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes() // No OTP
	b.Lac.EXPECT().GetDefaultReceive(ctx, gomock.Any(), gomock.Any()).Return(&linkedaccounts.LinkedAccount{ID: uuid.NewString(), WalletID: walletID}, nil).AnyTimes()

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

	assert.ErrorIs(t, ops.SetState(ctx, b, p.ID, payments.StateCompleted), payments.ErrInvalidStateTransition)
	assert.NoError(t, ops.SetState(ctx, b, p.ID, payments.StateConfirmed))
}

func TestGetRequiredActions(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	walletID := uuid.NewString()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Lac.EXPECT().ListBalances(ctx, gomock.Any()).Return([]linkedaccounts.LinkedAccount{}, nil).AnyTimes()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{WalletID: walletID, Provider: pti.ProviderName}, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
	})
	require.NoError(t, err)

	requiredActions, err := ops.GetRequiredActions(ctx, b, p.ID)
	require.NoError(t, err)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAccount)
	// assert.Contains(t, requiredActions, payments.RequiredActionTypeOTP)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeIPAddress)
}

func TestConfirm(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b.Tp = temporal_mock.NewMockClient(ctrl)
	b.Wc = wallets_mock.NewMockClient(ctrl)
	b.Lac = linkedaccounts_mock.NewMockClient(ctrl)
	b.Txc = transactions_mock.NewMockClient(ctrl)

	walletID := uuid.NewString()
	txID := uuid.NewString()
	b.Lac.EXPECT().ListBalances(ctx, gomock.Any()).Return([]linkedaccounts.LinkedAccount{}, nil).AnyTimes()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: pti.ProviderName, State: linkedaccounts.Verified, WalletID: walletID, SendCurrency: currency.USD, ReceiveCurrency: currency.USD}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{
		ID: walletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: walletID,
		},
		IPAddress: "193.9.4.6",
	})
	require.NoError(t, err)
	paymentID := p.ID

	_, requiredActions, err := ops.Confirm(ctx, b, paymentID)
	require.Error(t, err)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAmount)
	assert.Contains(t, requiredActions, payments.RequiredActionTypeSenderAccount)

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID: paymentID,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
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
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	walletID, receiverWalletID := uuid.NewString(), uuid.NewString()
	senderAccount := uuid.NewString()
	receiverAccount := uuid.NewString()
	newSenderAccount := uuid.NewString()
	newReceiverAccount := uuid.NewString()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
		ret := &linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: pti.ProviderName, State: linkedaccounts.Verified, WalletID: walletID, SendCurrency: currency.USD, ReceiveCurrency: currency.USD}
		if id == receiverAccount || id == newReceiverAccount {
			ret.WalletID = receiverWalletID
		}

		return ret, nil
	}).AnyTimes()
	b.Lac.EXPECT().Get(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{CanSend: true, CanReceive: true, Provider: pti.ProviderName, State: linkedaccounts.Verified, WalletID: walletID, SendCurrency: currency.USD, ReceiveCurrency: currency.USD}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().Get(ctx, receiverWalletID).Return(&wallets.Wallet{ID: receiverWalletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{
		ID: receiverWalletID,
	}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
	b.Txc.EXPECT().CountSendTransactions(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, walletID string) (int, error) {
		if walletID == "" {
			t.Fatal("walletID is empty for CountSendTransactions")
		}
		return 0, nil
	}).AnyTimes()

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
		SenderAccount:   senderAccount,
		ReceiverAccount: receiverAccount,
	})
	require.NoError(t, err)
	paymentID := p.ID

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID: paymentID,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
		},
		ReceiverAccount: receiverAccount,
	})
	require.NoError(t, err)
	assert.True(t, p.Receiver.IsEqual(payments.Identity{
		Type:       payments.IdentityTypeWalletURL,
		Identifier: "https://ilp.link/charlie",
	}))
	assert.True(t, p.SenderAmount.IsEqual(currency.FromFloat64(51, currency.USD)))
	assert.Equal(t, int64(5100), p.ReceiverAmount.Value)
	assert.Equal(t, senderAccount, p.SenderAccount)
	assert.Equal(t, receiverAccount, p.ReceiverAccount)

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID:              paymentID,
		SenderAmount:    currency.FromFloat64(54, currency.USD),
		SenderAccount:   newSenderAccount,
		ReceiverAccount: newReceiverAccount,
	})
	require.NoError(t, err)

	assert.True(t, p.Receiver.IsEqual(payments.Identity{
		Type:       payments.IdentityTypeWalletURL,
		Identifier: "https://ilp.link/charlie",
	}))

	assert.Equal(t, int64(5400), p.SenderAmount.Value)
	assert.Equal(t, int64(5400), p.ReceiverAmount.Value)
	assert.Equal(t, newReceiverAccount, p.ReceiverAccount)
	assert.Equal(t, newSenderAccount, p.SenderAccount)

	// update send amount
	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID:             paymentID,
		SenderAmount:   currency.FromFloat64(51, currency.USD),
		ReceiverAmount: currency.FromFloat64(51, currency.USD),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5100), p.SenderAmount.Value)

	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID:           paymentID,
		SenderAmount: currency.FromFloat64(55, currency.USD),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5500), p.SenderAmount.Value)
}

// crossProviderLinkedAccounts returns a lookup func for b.Lac.Get that resolves the given
// account IDs to the given linked accounts, erroring on any other ID.
func crossProviderLinkedAccounts(accs map[string]*linkedaccounts.LinkedAccount) func(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	return func(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
		if acc, ok := accs[id]; ok {
			return acc, nil
		}
		return nil, fmt.Errorf("unexpected linked account id %s", id)
	}
}

func TestCreate_CrossProvider(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b.Fc = features_mock.NewMockClient(ctrl)
	b.XagoClient = xago_mock.NewMockClient(ctrl)

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	// rafiki.ZARBalanceAccount is used as the sender account ID so that
	// validateSendBalances short-circuits instead of hitting the (unmocked) Xago client.
	senderAccountID := rafiki.ZARBalanceAccount
	receiverAccountID := uuid.NewString()

	b.Lac.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(crossProviderLinkedAccounts(map[string]*linkedaccounts.LinkedAccount{
		senderAccountID: {
			WalletID: senderWalletID, Provider: xago.ProviderName, Type: xago.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.ZAR, ReceiveCurrency: currency.ZAR,
		},
		receiverAccountID: {
			WalletID: receiverWalletID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.EUR, ReceiveCurrency: currency.EUR,
		},
	})).AnyTimes()
	b.Wc.EXPECT().Get(ctx, senderWalletID).Return(&wallets.Wallet{ID: senderWalletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{ID: receiverWalletID}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	b.Fc.EXPECT().Features(ctx, senderWalletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: true}, nil).AnyTimes()
	b.Fc.EXPECT().Features(ctx, receiverWalletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: true}, nil).AnyTimes()
	// Mocked Xago FX estimate for the Xago-Gatehub pair: 100 ZAR converts to 5 EUR at a rate of 0.05.
	b.XagoClient.EXPECT().EstimateConvertCurrency(ctx, xago_external.ZARtoEUR, gomock.Any()).Return(&xago_external.EstimateConvertCurrencyResponse{
		EstimatedRate:  json.Number("0.05"),
		ReceivedAmount: json.Number("5"),
	}, nil).AnyTimes()

	p, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: senderWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
		},
		SenderAmount:    currency.FromFloat64(100, currency.ZAR),
		SenderAccount:   senderAccountID,
		ReceiverAccount: receiverAccountID,
		IPAddress:       "193.9.4.6",
	})
	require.NoError(t, err)

	assert.Equal(t, senderAccountID, p.SenderAccount)
	assert.Equal(t, receiverAccountID, p.ReceiverAccount)
	// Receiver amount and FX rate reflect the mocked Xago FX conversion (100 ZAR -> 5 EUR, rate 0.05).
	assert.Equal(t, currency.EUR, p.ReceiverAmount.Currency)
	assert.Equal(t, int64(500), p.ReceiverAmount.Value)
	assert.Equal(t, 0.05, p.FXRate)
}

func TestCreate_CrossProvider_FeatureFlagDisabled(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b.Fc = features_mock.NewMockClient(ctrl)

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderAccountID := rafiki.ZARBalanceAccount
	receiverAccountID := uuid.NewString()

	b.Lac.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(crossProviderLinkedAccounts(map[string]*linkedaccounts.LinkedAccount{
		senderAccountID: {
			WalletID: senderWalletID, Provider: xago.ProviderName, Type: xago.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.ZAR, ReceiveCurrency: currency.ZAR,
		},
		receiverAccountID: {
			WalletID: receiverWalletID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.EUR, ReceiveCurrency: currency.EUR,
		},
	})).AnyTimes()
	b.Wc.EXPECT().Get(ctx, senderWalletID).Return(&wallets.Wallet{ID: senderWalletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{ID: receiverWalletID}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	b.Fc.EXPECT().Features(ctx, senderWalletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: true}, nil).AnyTimes()
	// Receiver hasn't opted in to Xago-Gatehub payments yet.
	b.Fc.EXPECT().Features(ctx, receiverWalletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: false}, nil).AnyTimes()

	_, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: senderWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
		},
		SenderAmount:    currency.FromFloat64(100, currency.ZAR),
		SenderAccount:   senderAccountID,
		ReceiverAccount: receiverAccountID,
		IPAddress:       "193.9.4.6",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
}

func TestCreate_CrossProvider_UnsupportedPair(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderAccountID := rafiki.ZARBalanceAccount
	receiverAccountID := uuid.NewString()

	// xago-pti isn't a supported cross provider pair (only xago-gatehub is registered).
	b.Lac.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(crossProviderLinkedAccounts(map[string]*linkedaccounts.LinkedAccount{
		senderAccountID: {
			WalletID: senderWalletID, Provider: xago.ProviderName, Type: xago.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.ZAR, ReceiveCurrency: currency.ZAR,
		},
		receiverAccountID: {
			WalletID: receiverWalletID, Provider: pti.ProviderName, Type: pti.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.USD, ReceiveCurrency: currency.USD,
		},
	})).AnyTimes()
	b.Wc.EXPECT().Get(ctx, senderWalletID).Return(&wallets.Wallet{ID: senderWalletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{ID: receiverWalletID}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	_, err := ops.Create(ctx, b, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: senderWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: "https://ilp.link/charlie",
		},
		SenderAmount:    currency.FromFloat64(100, currency.ZAR),
		SenderAccount:   senderAccountID,
		ReceiverAccount: receiverAccountID,
		IPAddress:       "193.9.4.6",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
}

func TestUpdate_CrossProvider(t *testing.T) {
	ctx, b := setupTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b.Fc = features_mock.NewMockClient(ctrl)
	b.XagoClient = xago_mock.NewMockClient(ctrl)

	walletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderAccount := uuid.NewString()
	receiverAccount := uuid.NewString()
	// New cross-provider accounts the payment gets moved to.
	xagoAccountID := rafiki.ZARBalanceAccount
	gatehubAccountID := uuid.NewString()

	b.Lac.EXPECT().Get(ctx, gomock.Any()).DoAndReturn(crossProviderLinkedAccounts(map[string]*linkedaccounts.LinkedAccount{
		senderAccount: {
			WalletID: walletID, Provider: pti.ProviderName, Type: pti.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.USD, ReceiveCurrency: currency.USD,
		},
		receiverAccount: {
			WalletID: receiverWalletID, Provider: pti.ProviderName, Type: pti.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.USD, ReceiveCurrency: currency.USD,
		},
		xagoAccountID: {
			WalletID: walletID, Provider: xago.ProviderName, Type: xago.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.ZAR, ReceiveCurrency: currency.ZAR,
		},
		gatehubAccountID: {
			WalletID: receiverWalletID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance,
			State: linkedaccounts.Verified, CanSend: true, CanReceive: true, SendCurrency: currency.EUR, ReceiveCurrency: currency.EUR,
		},
	})).AnyTimes()
	b.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{ID: walletID}, nil).AnyTimes()
	b.Wc.EXPECT().GetFromAddress(ctx, "https://ilp.link/charlie").Return(&wallets.Wallet{ID: receiverWalletID}, nil).AnyTimes()
	b.Txc.EXPECT().GetHasTransacted(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	b.Fc.EXPECT().Features(ctx, walletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: true}, nil).AnyTimes()
	b.Fc.EXPECT().Features(ctx, receiverWalletID).Return(&features.WalletFeatures{XagoGatehubPaymentsEnabled: true}, nil).AnyTimes()
	// Mocked Xago FX estimate for the Xago-Gatehub pair: 51 ZAR converts to 2.55 EUR at a rate of 0.05.
	b.XagoClient.EXPECT().EstimateConvertCurrency(ctx, xago_external.ZARtoEUR, gomock.Any()).Return(&xago_external.EstimateConvertCurrencyResponse{
		EstimatedRate:  json.Number("0.05"),
		ReceivedAmount: json.Number("2.55"),
	}, nil).AnyTimes()
	// validateSendBalances hits PTI for the original same-provider payment's balance check.
	b.Pti.EXPECT().GetBalance(ctx, gomock.Any()).Return(&pti.Balance{Available: currency.FromFloat64(1000, currency.USD), Total: currency.FromFloat64(1000, currency.USD)}, nil).AnyTimes()

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
		SenderAccount:   senderAccount,
		ReceiverAccount: receiverAccount,
		IPAddress:       "193.9.4.6",
	})
	require.NoError(t, err)
	paymentID := p.ID

	// Move the payment to a Xago-Gatehub cross-provider pair of accounts.
	p, err = ops.Update(ctx, b, payments.UpdateArgs{
		ID:              paymentID,
		SenderAccount:   xagoAccountID,
		ReceiverAccount: gatehubAccountID,
	})
	require.NoError(t, err)

	assert.Equal(t, xagoAccountID, p.SenderAccount)
	assert.Equal(t, gatehubAccountID, p.ReceiverAccount)
	assert.Equal(t, currency.ZAR, p.SenderAmount.Currency)
	// Receiver amount and FX rate reflect the mocked Xago FX conversion (51 ZAR -> 2.55 EUR, rate 0.05).
	assert.Equal(t, currency.EUR, p.ReceiverAmount.Currency)
	assert.Equal(t, int64(255), p.ReceiverAmount.Value)
	assert.Equal(t, 0.05, p.FXRate)
}

func TestSellerRisk(t *testing.T) {
	sellerSendTransactions := []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65}
	risk := []string{"0.0500", "0.0320", "0.0221", "0.0166", "0.0136", "0.0120", "0.0111", "0.0106", "0.0103", "0.0102", "0.0101", "0.0101", "0.0100", "0.0100"}

	for i, trxs := range sellerSendTransactions {
		sellerRisk := ops.SellerRisk(trxs)
		assert.Equal(t, fmt.Sprintf("%.4f", sellerRisk), risk[i])
	}
}
