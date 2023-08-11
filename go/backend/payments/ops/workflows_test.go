package ops_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateCardWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnActivity(a.SendPaymentSentEmail, mock.Anything, paymentID).Return(nil)
	env.OnActivity(a.SendPaymentReceivedEmail, mock.Anything, paymentID).Return(nil)
	env.OnActivity(a.SetPaymentState, mock.Anything, paymentID, payments.StateCompleted).Return(nil)

	env.ExecuteWorkflow(ops.PayinWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
}
