package deposits_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/deposits"
	_mx "gitlab.com/fynbos/backend/providers/mx"

	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env              *testsuite.TestWorkflowEnvironment
	depositsActivity *deposits.Activity
	mxActivity       *_mx.Activity
}

func (s *UnitTestSuite) SetupSuite() {
	s.depositsActivity = &deposits.Activity{}
	s.mxActivity = &_mx.Activity{}
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.depositsActivity)
	s.env.RegisterActivity(s.mxActivity)
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_DepositWorkflow_Success() {
	deposit := &deposits.Deposit{
		ID:              uuid.NewString(),
		AccountID:       uuid.NewString(),
		FundingSourceId: uuid.NewString(),
		Amount:          1000, // 10 USD
	}
	mxAcc := &_mx.Account{
		Guid:            uuid.NewString(),
		UserGuid:        uuid.NewString(),
		MemberGuid:      uuid.NewString(),
		FundingsourceID: deposit.FundingSourceId,
	}
	trxId := uuid.New()
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})
	s.env.OnActivity(s.depositsActivity.GetDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, depositID string) (*deposits.Deposit, error) {
			s.Equal(deposit.ID, depositID)
			return deposit, nil
		},
	)
	s.env.OnActivity(s.mxActivity.GetMxAccountByFundingsource, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, fundingsourceID string) (*_mx.Account, error) {
			s.Equal(deposit.FundingSourceId, fundingsourceID)
			return mxAcc, nil
		},
	)
	s.env.OnActivity(s.mxActivity.StartBalanceAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, mxAccountGuid string) error {
			s.Equal(mxAcc.Guid, mxAccountGuid)
			return nil
		},
	)
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, mxAccountGuid string, maxRetries uint8, pollInterval time.Duration) error {
			s.Equal(mxAcc.Guid, mxAccountGuid)
			return nil
		},
	)
	s.env.OnActivity(s.mxActivity.GetMxAccountBalance, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, mxAccountGuid string) (*_mx.AccountBalance, error) {
			s.Equal(mxAcc.Guid, mxAccountGuid)
			return &_mx.AccountBalance{
				AssetCode:  "USD",
				AssetScale: 2,
				Value:      1500,
			}, nil
		},
	)
	s.env.OnActivity(s.depositsActivity.CreatePendingDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (string, error) {
			s.Equal(deposit.ID, id)
			return trxId.String(), nil
		})
	s.env.OnActivity(s.depositsActivity.ProcessNoopDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(deposit.ID, id)
			return nil
		})
	s.env.OnActivity(s.depositsActivity.PostPendingDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) error {
			s.Equal(trxId.String(), id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TODO add more comprehensive testing for branching in workflow and rollbacks
// see https://docs.temporal.io/docs/go/how-to-test-workflow-definitions-in-go/

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
