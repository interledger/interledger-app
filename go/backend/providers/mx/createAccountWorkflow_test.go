package mx

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type MxAccountTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env      *testsuite.TestWorkflowEnvironment
	activity *Activity
}

func (s *MxAccountTestSuite) SetupSuite() {
	s.activity = &Activity{}
}

func (s *MxAccountTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.activity)
}

func (s *MxAccountTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *MxAccountTestSuite) TestCreateMxAccountSuccess() {
	userID := uuid.NewString()
	accountID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	fundingsourceName := "todo"
	mxAccountGuid := "acct_" + uuid.NewString()
	mxMemberGuid := "mbr_" + uuid.NewString()
	mxUserGuid := "usr_" + uuid.NewString()

	s.env.OnActivity(s.activity.GetSelectedMxAccountGuid, mock.Anything, mxUserGuid, mxMemberGuid).Return(
		func(ctx context.Context, userGuid string, memberGuid string) (string, error) {
			s.Equal(mxUserGuid, userGuid)
			s.Equal(mxMemberGuid, memberGuid)
			return mxAccountGuid, nil
		},
	)

	s.env.OnActivity(
		s.activity.CreateMxAccount,
		mock.Anything,
		fundingsourceID,
		accountID,
		mxUserGuid,
		mxMemberGuid,
		mxAccountGuid,
	).Return(nil)

	s.env.OnActivity(
		s.activity.StartIdentityAggregation,
		mock.Anything,
		mxAccountGuid,
	).Return(nil)

	s.env.OnActivity(
		s.activity.WaitForAggregation,
		mock.Anything,
		mock.Anything,
	).Return(func(ctx context.Context, args *WaitForAggregationArgs) error {
		s.Equal(mxAccountGuid, args.MxAccountGuid)
		s.Equal(uint8(5), args.MaxRetries)
		s.Equal(12*time.Second, args.PollInterval)
		return nil
	})

	s.env.OnActivity(s.activity.VerifyOwnership, mock.Anything, mxAccountGuid).Return(nil)

	s.env.OnActivity(s.activity.CreateUnitCounterParty, mock.Anything, mxAccountGuid).Return(nil)

	s.env.OnActivity(
		s.activity.CreateFundingSource,
		mock.Anything,
		mxAccountGuid,
		fundingsourceName,
	).Return(nil)

	s.env.ExecuteWorkflow(CreateMxAccountWorkflow, &CreateMxAccountWorkflowArgs{
		ID:         fundingsourceID,
		UserGuid:   mxUserGuid,
		MemberGuid: mxMemberGuid,
		AccountID:  accountID,
		IdentityID: userID,
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/
func TestCreateMxAccountWorkflow(t *testing.T) {
	suite.Run(t, new(MxAccountTestSuite))
}
