package mx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gitlab.com/fynbos/backend/providers/mx"
	"go.temporal.io/sdk/testsuite"
)

type MxFundingsourceTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env      *testsuite.TestWorkflowEnvironment
	activity *mx.Activity
}

func (s *MxFundingsourceTestSuite) SetupSuite() {
	s.activity = &mx.Activity{}
}

func (s *MxFundingsourceTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.activity)
}

func (s *MxFundingsourceTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *MxFundingsourceTestSuite) TestSuccess() {
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	name := "test mx account"
	mxAccount := &mx.Account{
		Guid:            "acct_" + uuid.NewString(),
		MemberGuid:      "mbr_" + uuid.NewString(),
		UserGuid:        "usr_" + uuid.NewString(),
		AccountID:       accountID,
		FundingsourceID: fundingsourceID,
	}

	s.env.OnActivity(s.activity.CreateUnitCounterParty, mock.Anything, mxAccount.Guid).Return(nil)

	s.env.OnActivity(
		s.activity.CreateFundingSource,
		mock.Anything,
		mxAccount.Guid,
		name,
	).Return(nil)

	s.env.ExecuteWorkflow(mx.CreateMxAccountWorkflow, &mx.MxCreateFundingsourceWorkflowArgs{
		MxAccountGuid: mxAccount.Guid,
		AccountID:     accountID,
		Name:          name,
	})
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/
func TestCreateMxFundingsourceWorkflow(t *testing.T) {
	suite.Run(t, new(mx.MxAccountTestSuite))
}
