package workflows

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/providers"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/transactions"

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
	mArgs := ProviderWorkflowArgs{
		Args: providers.TransfersArgs{
			FromForeignID:       uuid.NewString(),
			ToForeignID:         uuid.NewString(),
			FromPaymentPointer:  uuid.NewString(),
			ToPaymentPointer:    uuid.NewString(),
			FromLinkedAccountID: uuid.NewString(),
			ToLinkedAccountID:   uuid.NewString(),
			FromWalletID:        uuid.NewString(),
			ToWalletID:          uuid.NewString(),
			Amount:              amt,
			FromTransactionID:   trxID,
			IPAddress:           "198.0.0.4",
		},
		Key: providers.GMTACH2ACH,
	}

	env.OnActivity(a.GetProviderWorkflowArgs, mock.Anything, id).Return(&mArgs, nil)
	env.OnActivity(a.AddContact, mock.Anything, mArgs.Args.FromPaymentPointer, mArgs.Args.ToPaymentPointer).Return(nil)
	env.OnWorkflow(gmt_workflows.ACH2ACHTransferWorkflow, mock.Anything, mArgs.Args).Return(&providers.TransferResponse{
		Type:                       providers.GMTACH2ACH,
		OutgoingTransferState:      transactions.StateCompleted,
		OutgoingTransferExternalID: "external_id",
		IncomingTransferState:      transactions.StateCompleted,
		IncomingTransferExternalID: "external_id",
	}, nil)
	env.OnActivity(a.CompleteOutgoingPayment, mock.Anything, id, mock.Anything).Return(nil)
	env.OnActivity(a.SendOutgoingPaymentReceipt, mock.Anything, id, mock.Anything).Return(nil)
	env.OnActivity(a.SendIncomingPaymentReceipt, mock.Anything, id).Return(nil)

	env.RegisterWorkflow(gmt_workflows.ACH2ACHTransferWorkflow)
	env.ExecuteWorkflow(OutgoingTransactionWorkflow, id, trxID, "198.0.0.4", "")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
}

func TestOutgoingTransactionSendsFailedTransactionEmail(t *testing.T) {
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

	mArgs := ProviderWorkflowArgs{
		Args: providers.TransfersArgs{
			FromForeignID:       uuid.NewString(),
			ToForeignID:         uuid.NewString(),
			FromPaymentPointer:  uuid.NewString(),
			ToPaymentPointer:    uuid.NewString(),
			FromLinkedAccountID: uuid.NewString(),
			ToLinkedAccountID:   uuid.NewString(),
			FromWalletID:        uuid.NewString(),
			ToWalletID:          uuid.NewString(),
			Amount:              currency.FromFloat64(10.5, currency.ParseCurrency("USD")),
			FromTransactionID:   trxID,
			IPAddress:           "198.0.0.3",
		},
		Key: providers.GMTACH2ACH,
	}

	env.OnActivity(a.GetProviderWorkflowArgs, mock.Anything, id).Return(&mArgs, nil)
	env.OnActivity(a.AddContact, mock.Anything, mArgs.Args.FromPaymentPointer, mArgs.Args.ToPaymentPointer).Return(nil)
	env.OnWorkflow(gmt_workflows.ACH2ACHTransferWorkflow, mock.Anything, mArgs.Args).Return(&providers.TransferResponse{
		Type:                       providers.GMTACH2ACH,
		OutgoingTransferState:      transactions.StateFailed,
		OutgoingTransferExternalID: "external_id",
		IncomingTransferState:      transactions.StateFailed,
		IncomingTransferExternalID: "external_id",
	}, nil)
	env.OnActivity(a.FailOutgoingPayment, mock.Anything, id).Return(nil)
	env.OnActivity(a.SendFailedOutgoingPaymentMail, mock.Anything, id).Return(nil)

	env.RegisterWorkflow(gmt_workflows.ACH2ACHTransferWorkflow)
	env.ExecuteWorkflow(OutgoingTransactionWorkflow, id, trxID, "198.0.0.3", "")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "", result)
	env.AssertExpectations(t)
}
