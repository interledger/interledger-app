package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client"
	"gitlab.com/fynbos/backend/payments/ops"
	gmt_ops "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/providers/tabapay"
	tabapay_mock "gitlab.com/fynbos/backend/providers/tabapay/client/mock"
	tabapay_external "gitlab.com/fynbos/backend/providers/tabapay/external"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"gotest.tools/assert"
)

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	em := email_mock.NewMockClient(ctrl)
	em.EXPECT().SendPaymentReceivedEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentSentEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentFailedEmail(gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendConnectedAccountEmail(gomock.Any(), gomock.Any()).AnyTimes()
	tc := tabapay_mock.NewMockClient(ctrl)
	tc.EXPECT().PullFromCard(gomock.Any(), gomock.Any()).Return(
		&tabapay.Transaction{
			ID:        uuid.NewString(),
			Status:    string(tabapay_external.TransactionStatusCompleted),
			NetworkRC: "00",
		},
		nil,
	).AnyTimes()
	tc.EXPECT().PushToCard(gomock.Any(), gomock.Any()).Return(
		&tabapay.Transaction{
			ID:        uuid.NewString(),
			Status:    string(tabapay_external.TransactionStatusCompleted),
			NetworkRC: "00",
		},
		nil,
	).AnyTimes()
	tc.EXPECT().Get3DSSession(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, id string) (*tabapay.ThreeDSSession, error) {
			return &tabapay.ThreeDSSession{ID: id, ECI: tabapay.ThreeDSFullyAuthenticated}, nil
		},
	).AnyTimes()

	b := &TestBackends{
		db:      db.MigrateTestDB(t, ctx),
		tabapay: tc,
		user:    user_mock.NewMock(),
		email:   em,
	}
	sendWalletID, sendLinkedAccount := createTestWallet(t, b)
	receiveWalletID, receiveLinkedAccount := createTestWallet(t, b)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(ops.NewActivity(b))
	env.RegisterWorkflow(ops.PaymentWorkflow)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	env.RegisterWorkflow(gmt_ops.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_ops.GMTNotifyCompleted)

	tp := temporal_mock.NewMockClient(ctrl)
	tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(arg1 interface{}, arg2 interface{}, arg3 interface{}, arg4 ...interface{}) (*workflow.Execution, error) {
		require.Len(t, arg4, 1)
		env.ExecuteWorkflow(ops.PaymentWorkflow, arg4[0].(string))

		return nil, nil
	})
	b.temporal = tp

	pc := client.New(b)
	p, err := pc.Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: sendWalletID,
		},
		SenderAccount: sendLinkedAccount,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: receiveWalletID,
		},
		ReceiverAccount: receiveLinkedAccount,
		SenderAmount:    currency.FromUInt64(10, currency.ParseCurrency("USD")),
		ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("USD")),
	})
	require.NoError(t, err)

	p, err = pc.Update(ctx, payments.UpdateArgs{
		ID:        p.ID,
		ThreeDSID: "123",
	})
	require.NoError(t, err)

	p, requiredActions, err := pc.Confirm(ctx, p.ID)
	require.NoError(t, err)
	require.Empty(t, requiredActions)

	for {
		if env.IsWorkflowCompleted() {
			break
		}
	}
	require.NoError(t, env.GetWorkflowError())

	payment, err := pc.Lookup(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, payments.StateCompleted, payment.State)
}

/*
Seeds a user:
- user client returns user for userID
- user client returns list of users
- wallet is created for userID
- linked card is created
- tabapay account is created
*/
func createTestWallet(t *testing.T, b *TestBackends) (string, string) {
	userID := uuid.NewString()
	address, err := wallets.ParseAddress(fmt.Sprintf("https://fynbos.test/%s", faker.FirstName()))
	if err != nil {
		t.Fatal(err)
	}

	wallet, err := b.Wallets().Create(context.Background(), wallets.CreateArgs{
		UserID:    userID,
		Addresses: []wallets.Address{address},
	})
	if err != nil {
		t.Fatal(err)
	}
	b.user.MapUserWallet(context.Background(), userID, wallet.ID)

	la, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
		WalletID:   wallet.ID,
		Name:       "default",
		Provider:   tabapay.ProviderName,
		ProviderID: uuid.NewString(),
		CanSend:    true,
		CanReceive: true,
		Type:       tabapay.TypeCard,
	})
	if err != nil {
		t.Fatal(err)
	}

	return wallet.ID, la.ID
}
