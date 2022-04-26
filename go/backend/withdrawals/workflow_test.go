package withdrawals_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/withdrawals"
	"testing"

	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env      *testsuite.TestWorkflowEnvironment
	activity *withdrawals.Activity
}

func (s *UnitTestSuite) SetupSuite() {
	activity := withdrawals.Activity{}

	s.activity = &activity
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.activity)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_WithdrawalWorkflow_Success() {
	withdrawalId := uuid.New()
	trxId := uuid.New()
	s.env.OnActivity(s.activity.SetWithdrawalState, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, state withdrawals.State) error {
			s.Equal(withdrawalId.String(), id)
			return nil
		})
	s.env.OnActivity(s.activity.CreatePendingWithdrawalTransaction, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (string, error) {
			s.Equal(withdrawalId.String(), id)
			return trxId.String(), nil
		})
	s.env.OnActivity(s.activity.ProcessNoopWithdrawal, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(withdrawalId.String(), id)
			return nil
		})
	s.env.OnActivity(s.activity.PostPendingWithdrawalTransaction, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(trxId.String(), id)
			return nil
		})

	s.env.ExecuteWorkflow(withdrawals.WithdrawalWorkflow, withdrawalId.String())

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
