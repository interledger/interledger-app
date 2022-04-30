package deposits_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/deposits"
	"testing"

	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
	da  *deposits.Activity
}

func (s *UnitTestSuite) SetupSuite() {
	da := deposits.Activity{}

	s.da = &da
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.da)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_DepositWorkflow_Success() {
	depositId := uuid.New()
	trxId := uuid.New()
	s.env.OnActivity(s.da.SetDepositState, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(depositId.String(), id)
			return nil
		})
	s.env.OnActivity(s.da.CreatePendingDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (string, error) {
			s.Equal(depositId.String(), id)
			return trxId.String(), nil
		})
	s.env.OnActivity(s.da.ProcessNoopDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(depositId.String(), id)
			return nil
		})
	s.env.OnActivity(s.da.PostPendingDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(trxId.String(), id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, depositId.String())

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
