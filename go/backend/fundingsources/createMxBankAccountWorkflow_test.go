package fundingsources

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env      *testsuite.TestWorkflowEnvironment
	activity *Activity
}

func (s *UnitTestSuite) SetupSuite() {
	pa := Activity{}

	s.activity = &pa
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.activity)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_Outgoing_Payment_Workflow_Success() {
	userID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	mxUserGuid := uuid.NewString()
	mxMemberGuid := uuid.NewString()
	mxAccountGuid := uuid.NewString()

	s.env.OnActivity(s.activity.GetSelectedMxAccountGuid, mock.Anything, mxUserGuid, mxMemberGuid).Return(
		func(ctx context.Context, userGuid string, memberGuid string) (string, error) {
			s.Equal(mxUserGuid, userGuid)
			s.Equal(mxMemberGuid, memberGuid)
			return mxAccountGuid, nil
		})

	s.env.OnActivity(s.activity.CreateMxAccount, mock.Anything, fundingsourceID, mxUserGuid, mxMemberGuid, mxAccountGuid).Return(
		func(ctx context.Context, fsID string, userGuid string, memberGuid string, accountGuid string) error {
			s.Equal(fundingsourceID, fsID)
			s.Equal(mxUserGuid, userGuid)
			s.Equal(mxMemberGuid, memberGuid)
			s.Equal(mxAccountGuid, accountGuid)
			return nil
		})

	s.env.OnActivity(s.activity.StartIdentityAggregation, mock.Anything, fundingsourceID).Return(
		func(ctx context.Context, fsID string) error {
			s.Equal(fundingsourceID, fsID)
			return nil
		})

	s.env.OnActivity(s.activity.WaitForIdentityAggregation, mock.Anything, fundingsourceID).Return(
		func(ctx context.Context, fsID string) error {
			s.Equal(fundingsourceID, fsID)
			return nil
		})

	s.env.OnActivity(s.activity.VerifyOwnership, mock.Anything, fundingsourceID, userID).Return(
		func(ctx context.Context, fsID string, identityID string) error {
			s.Equal(fundingsourceID, fsID)
			s.Equal(userID, identityID)
			return nil
		})

	s.env.OnActivity(s.activity.SetMask, mock.Anything, fundingsourceID).Return(
		func(ctx context.Context, fsID string) error {
			s.Equal(fundingsourceID, fsID)
			return nil
		})

	s.env.OnActivity(s.activity.CreateUnitCounterParty, mock.Anything, fundingsourceID).Return(
		func(ctx context.Context, fsID string) error {
			s.Equal(fundingsourceID, fsID)
			return nil
		})

	s.env.ExecuteWorkflow(CreateMxBankAccountWorkflow, &CreateMxBankAccountWorkflowArgs{
		FundingSourceID: fundingsourceID,
		MxUserGuid:      mxUserGuid,
		MxMemberGuid:    mxMemberGuid,
		IdentityID:      userID,
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/

func TestCreateMxAccountWorkflow(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
