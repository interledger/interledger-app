package ops_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/payments/ops"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func assertPaymentID(t *testing.T, id string) func(ctx context.Context, id string) error {
	return func(ctx context.Context, paymentID string) error {
		info := activity.GetInfo(ctx)
		assert.Equal(t, id, paymentID, fmt.Sprintf("%s called with incorrect paymentID", info.ActivityType.Name))
		return nil
	}
}

func assertWorkflowPaymentID(t *testing.T, id string) func(ctx workflow.Context, id string) error {
	return func(ctx workflow.Context, paymentID string) error {
		info := workflow.GetInfo(ctx)
		assert.Equal(t, id, paymentID, fmt.Sprintf("%s called with incorrect paymentID", info.WorkflowType.Name))
		return nil
	}
}

func TestCreatePaymentWorkflowGoldenPath(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.CheckPaymentSuccess, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(gmt_workflows.GMTNotifyCompleted, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 0)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 1)
}

func TestCreatePaymentWorkflowComplianceCheckFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(errors.New("Failed compliance"))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.UpdatePayInTransactionState, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(gmt_workflows.GMTNotifyCompleted, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "PayinWorkflow", 0)
	env.AssertNumberOfCalls(t, "PayoutWorkflow", 0)
	env.AssertNumberOfCalls(t, "GMTNotifyCompleted", 0)
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 1)
	env.AssertNumberOfCalls(t, "UpdatePayInTransactionState", 1)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 0)
}

func TestCreatePaymentWorkflowPayinFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(errors.New("Pay in failure"))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(gmt_workflows.GMTNotifyCompleted, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "GMTNotifyCompleted", 0)
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 1)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 0)
}

func TestCreatePaymentWorkflowPayoutFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(errors.New("Pay out failure"))
	env.OnWorkflow(gmt_workflows.GMTNotifyCompleted, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "GMTNotifyCompleted", 0)
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 1)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 0)
}

func TestCreatePaymentWorkflowNotifyGmtCompletedFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(gmt_workflows.GMTNotifyCompleted, mock.Anything, mock.Anything).Return(errors.New("Failed to notify GMT"))
	env.OnActivity(a.CheckPaymentSuccess, mock.Anything, mock.Anything).Return(true, nil)
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 0)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 1)
}

func TestCreatePaymentWorkflowPaymentFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(gmt_workflows.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_workflows.GMTNotifyCompleted)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	b := ops.NewTestBackends(t)
	a := ops.NewActivity(b)

	paymentID := uuid.NewString()
	env.OnWorkflow(gmt_workflows.GMTComplianceChecksWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateProcessing, mock.Anything, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayinWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnWorkflow(ops.PayoutWorkflow, mock.Anything, mock.Anything).Return(assertWorkflowPaymentID(t, paymentID))
	env.OnActivity(a.CheckPaymentSuccess, mock.Anything, mock.Anything).Return(false, nil)
	env.OnActivity(a.SetPaymentStateFailed, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))
	env.OnActivity(a.SetPaymentStateComplete, mock.Anything, mock.Anything).Return(assertPaymentID(t, paymentID))

	env.ExecuteWorkflow(ops.PaymentWorkflow, paymentID)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertNumberOfCalls(t, "SetPaymentStateFailed", 1)
	env.AssertNumberOfCalls(t, "SetPaymentStateComplete", 0)
}
