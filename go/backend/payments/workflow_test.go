package payments_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/payments"
	"testing"

	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
	pa  *payments.Activity
}

func (s *UnitTestSuite) SetupSuite() {
	pa := payments.Activity{}

	s.pa = &pa
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.pa)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_Outgoing_Payment_Workflow_Success() {
	paymentId := uuid.New()
	trxId := uuid.New()
	s.env.OnActivity(s.pa.SetOutgoingPaymentState, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, state payments.State) error {
			s.Equal(paymentId.String(), id)
			return nil
		})
	s.env.OnActivity(s.pa.CreatePendingOutgoingPayment, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (string, error) {
			s.Equal(paymentId.String(), id)
			return trxId.String(), nil
		})
	s.env.OnActivity(s.pa.ProcessNoopOutgoingPayment, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(paymentId.String(), id)
			return nil
		})
	s.env.OnActivity(s.pa.PostPendingOutgoingPayment, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(trxId.String(), id)
			return nil
		})

	s.env.ExecuteWorkflow(payments.OutgoingPaymentWorkflow, paymentId.String())

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
