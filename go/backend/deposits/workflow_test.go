package deposits_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/fynbos/backend/deposits"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/unit"

	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/testsuite"
)

type AchDepositTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env              *testsuite.TestWorkflowEnvironment
	depositsActivity *deposits.Activity
	mxActivity       *_mx.Activity
	unitActivity     *unit.Activity
}

func (s *AchDepositTestSuite) SetupSuite() {
	s.depositsActivity = &deposits.Activity{}
	s.mxActivity = &_mx.Activity{}
	s.unitActivity = &unit.Activity{}
}

func (s *AchDepositTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterActivity(s.depositsActivity)
	s.env.RegisterActivity(s.mxActivity)
	s.env.RegisterActivity(s.unitActivity)
}

func (s *AchDepositTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *AchDepositTestSuite) TestGoldenPath() {
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
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
	s.env.OnActivity(s.unitActivity.UnitInitiateUserDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.InitiateUserDepositArgs) (*unit.UserAchDeposit, error) {
			s.Equal(deposit.ID, args.DepositID)
			s.Equal(deposit.FundingSourceId, args.FundingsourceID)
			s.Equal(deposit.AccountID, args.AccountID)
			s.Equal(deposit.Amount, args.Amount)
			s.Equal("Fynbos", args.Description)
			return nil, nil
		})

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unit-user-ach-deposit", string(unit.PAYMENT_SENT))
	}, time.Millisecond*1)

	s.env.OnActivity(s.depositsActivity.CreateAchDepositTransactions, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id, transferID string) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *AchDepositTestSuite) TestFailsDespositOnInsufficientBalance() {
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
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Processing).Return(
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
			return nil
		},
	)
	s.env.OnActivity(s.mxActivity.GetMxAccountBalance, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, mxAccountGuid string) (*_mx.AccountBalance, error) {
			s.Equal(mxAcc.Guid, mxAccountGuid)
			return &_mx.AccountBalance{
				AssetCode:  "USD",
				AssetScale: 2,
				Value:      900,
			}, nil
		},
	)
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Failed).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.ErrorContains(s.env.GetWorkflowError(), "Insufficient funding source balance")
}

func (s *AchDepositTestSuite) TestFailsDespositIfFundingsourceIsNotInUSD() {
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
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Processing).Return(
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
			return nil
		},
	)
	s.env.OnActivity(s.mxActivity.GetMxAccountBalance, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, mxAccountGuid string) (*_mx.AccountBalance, error) {
			s.Equal(mxAcc.Guid, mxAccountGuid)
			return &_mx.AccountBalance{
				AssetCode:  "ZAR",
				AssetScale: 2,
				Value:      1100,
			}, nil
		},
	)
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Failed).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.ErrorContains(s.env.GetWorkflowError(), "account is not in USD")
}

func (s *AchDepositTestSuite) TestFailsDespositIfAchIsRejected() {
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

	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Processing).Return(
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
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
	s.env.OnActivity(s.unitActivity.UnitInitiateUserDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.InitiateUserDepositArgs) (*unit.UserAchDeposit, error) {
			s.Equal(deposit.ID, args.DepositID)
			s.Equal(deposit.FundingSourceId, args.FundingsourceID)
			s.Equal(deposit.AccountID, args.AccountID)
			s.Equal(deposit.Amount, args.Amount)
			s.Equal("Fynbos", args.Description)
			return nil, nil
		})
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unit-user-ach-deposit", string(unit.PAYMENT_REJECTED))
	}, time.Millisecond*1)
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Failed).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.ErrorContains(s.env.GetWorkflowError(), "ACH failed. achStatus="+string(unit.PAYMENT_REJECTED))
}

func (s *AchDepositTestSuite) TestFailsDespositIfAchIsReturned() {
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

	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Processing).Return(
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
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
	s.env.OnActivity(s.unitActivity.UnitInitiateUserDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.InitiateUserDepositArgs) (*unit.UserAchDeposit, error) {
			s.Equal(deposit.ID, args.DepositID)
			s.Equal(deposit.FundingSourceId, args.FundingsourceID)
			s.Equal(deposit.AccountID, args.AccountID)
			s.Equal(deposit.Amount, args.Amount)
			s.Equal("Fynbos", args.Description)
			return nil, nil
		})
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unit-user-ach-deposit", string(unit.PAYMENT_RETURNED))
	}, time.Millisecond*1)
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Failed).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.ErrorContains(s.env.GetWorkflowError(), "ACH failed. achStatus="+string(unit.PAYMENT_RETURNED))
}

func (s *AchDepositTestSuite) TestFailsDespositIfAchIsCanceled() {
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

	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Processing).Return(
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
	s.env.OnActivity(s.mxActivity.WaitForAggregation, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *_mx.WaitForAggregationArgs) error {
			s.Equal(mxAcc.Guid, args.MxAccountGuid)
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
	s.env.OnActivity(s.unitActivity.UnitInitiateUserDeposit, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, args *unit.InitiateUserDepositArgs) (*unit.UserAchDeposit, error) {
			s.Equal(deposit.ID, args.DepositID)
			s.Equal(deposit.FundingSourceId, args.FundingsourceID)
			s.Equal(deposit.AccountID, args.AccountID)
			s.Equal(deposit.Amount, args.Amount)
			s.Equal("Fynbos", args.Description)
			return nil, nil
		})
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unit-user-ach-deposit", string(unit.PAYMENT_CANCELED))
	}, time.Millisecond*1)
	s.env.OnActivity(s.depositsActivity.SetDepositState, mock.Anything, mock.Anything, deposits.Failed).Return(
		func(ctx context.Context, id string, state deposits.State) error {
			s.Equal(deposit.ID, id)
			return nil
		})

	s.env.ExecuteWorkflow(deposits.DepositWorkflow, deposit.ID)

	s.True(s.env.IsWorkflowCompleted())
	s.ErrorContains(s.env.GetWorkflowError(), "ACH failed. achStatus="+string(unit.PAYMENT_CANCELED))
}

func TestDepositWorkflow(t *testing.T) {
	suite.Run(t, new(AchDepositTestSuite))
}
