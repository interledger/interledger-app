package workflows

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/db"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"

	"gitlab.com/fynbos/backend/providers/machnet"

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
	b := NewTestBackends(t,
		db.MigrateTestDB(t, context.Background()),
		nil,
		la_mock, nil)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(machnet_workflows.CreateTransactionWorkflow)

	a := NewActivity(b)

	id := uuid.NewString()

	mArgs := machnet.CreateTransactionArgs{
		FromForeignID:       uuid.NewString(),
		ToForeignID:         uuid.NewString(),
		FromLinkedAccountID: uuid.NewString(),
		ToLinkedAccountID:   uuid.NewString(),
		Amount:              10.5,
		Currency:            "USD",
		IPAddress:           "198.0.0.3",
	}

	env.OnActivity(a.GetProviderArgs, mock.Anything, id).Return(&mArgs, nil)
	env.OnWorkflow(machnet_workflows.CreateTransactionWorkflow, mock.Anything, mArgs).Return("external_id", nil)
	env.OnActivity(a.CompleteOutgoingPayment, mock.Anything, id, "external_id").Return(nil)
	env.OnActivity(a.SendOutgoingPaymentReceipt, mock.Anything, id, "external_id").Return(nil)

	env.ExecuteWorkflow(OutgoingTransactionWorkflow, id, "198.0.0.3")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "external_id", result)
}
