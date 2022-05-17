package onboarding

import (
	context "context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
	oa  *Activity
}

func (s *UnitTestSuite) SetupSuite() {
	s.oa = &Activity{}
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.oa)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) TestOnboardingSuccess() {
	accountID := uuid.NewString()
	workflowArgs := &OnboardUnitCustomerArgs{
		CustomerID: uuid.NewString(),
		Type:       "individual",
		IdentityID: uuid.NewString(),
		Country:    "US",
	}
	s.env.OnActivity(s.oa.CreateAccount, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, identityID string, country string) (string, error) {
			s.Equal(workflowArgs.IdentityID, identityID)
			s.Equal(workflowArgs.Country, country)
			return accountID, nil
		},
	)
	s.env.OnActivity(s.oa.MapCustomerToAccount, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *MapCustomerToAccountArgs) error {
			s.Equal(workflowArgs.Type, args.Type)
			s.Equal(workflowArgs.CustomerID, args.CustomerID)
			s.Equal(accountID, args.AccountID)
			return nil
		},
	)
	s.env.ExecuteWorkflow(OnboardUnitCustomerWorkflow, workflowArgs)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
