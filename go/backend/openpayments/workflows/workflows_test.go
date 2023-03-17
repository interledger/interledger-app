package workflows

import (
	"context"
	"gitlab.com/fynbos/backend/providers/machnet"
	"testing"

	"gitlab.com/fynbos/backend/currency"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"

	"gitlab.com/fynbos/backend/db"

	"github.com/google/uuid"

	linked_account_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestOutgoingTransactionWorkflow(t *testing.T) {

	ctrl := gomock.NewController(t)

	la_mock := linked_account_mock.NewMockClient(ctrl)
	kyc_mock := kyc_mock.NewMockClient(ctrl)
	b := NewTestBackends(t,
		db.MigrateTestDB(t, context.Background()),
		nil,
		la_mock, nil, kyc_mock, nil)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	id := uuid.NewString()
	trxID := uuid.NewString()

	amt := currency.FromFloat64(10.55, currency.USD)
	mArgs := machnet.CreateTransactionArgs{
		FromForeignID:       uuid.NewString(),
		ToForeignID:         uuid.NewString(),
		FromPaymentPointer:  uuid.NewString(),
		ToPaymentPointer:    uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		ToLinkedAccountID:   uuid.NewString(),
		Amount:              amt,
		IPAddress:           "198.0.0.4",
	}

	env.OnActivity(a.GetProviderArgs, mock.Anything, id).Return(&mArgs, nil)
	env.OnActivity(a.AddContact, mock.Anything, mArgs.FromPaymentPointer, mArgs.ToPaymentPointer).Return(nil)
	env.OnActivity(a.MarkContactLastPaid, mock.Anything, mArgs.FromPaymentPointer, mArgs.ToPaymentPointer).Return(nil)

	env.OnActivity(a.CompleteOutgoingPayment, mock.Anything, id, mock.Anything).Return(nil)
	env.OnActivity(a.SendOutgoingPaymentReceipt, mock.Anything, id, mock.Anything).Return(nil)
	env.OnActivity(a.SendIncomingPaymentReceipt, mock.Anything, id).Return(nil)

	env.ExecuteWorkflow(OutgoingTransactionWorkflow, id, trxID, "198.0.0.4")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
}

func TestOutgoingTransactionSendsFailedTransactionEmail(t *testing.T) {
	t.Skip("Skip until trx are back")
	ctrl := gomock.NewController(t)

	la_mock := linked_account_mock.NewMockClient(ctrl)
	kyc_mock := kyc_mock.NewMockClient(ctrl)
	b := NewTestBackends(t,
		db.MigrateTestDB(t, context.Background()),
		nil,
		la_mock, nil, kyc_mock, nil)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivity(b)

	id := uuid.NewString()
	trxID := uuid.NewString()

	mArgs := machnet.CreateTransactionArgs{
		FromForeignID:       uuid.NewString(),
		ToForeignID:         uuid.NewString(),
		FromPaymentPointer:  uuid.NewString(),
		ToPaymentPointer:    uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		ToLinkedAccountID:   uuid.NewString(),
		Amount:              currency.FromFloat64(10.5, currency.ParseCurrency("USD")),
		IPAddress:           "198.0.0.3",
	}

	env.OnActivity(a.GetProviderArgs, mock.Anything, id).Return(&mArgs, nil)
	env.OnActivity(a.AddContact, mock.Anything, mArgs.FromPaymentPointer, mArgs.ToPaymentPointer).Return(nil)
	env.OnActivity(a.MarkContactLastPaid, mock.Anything, mArgs.FromPaymentPointer, mArgs.ToPaymentPointer).Return(nil)
	env.OnActivity(a.FailOutgoingPayment, mock.Anything, id).Return(nil)
	env.OnActivity(a.SendFailedOutgoingPaymentMail, mock.Anything, id).Return(nil)

	env.ExecuteWorkflow(OutgoingTransactionWorkflow, id, trxID, "198.0.0.3")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "", result)
	env.AssertExpectations(t)
}
