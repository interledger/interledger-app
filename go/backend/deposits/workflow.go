package deposits

import (
	"fmt"
	"time"

	_mx "gitlab.com/fynbos/backend/providers/mx"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/*
 - Start balance aggregation
    If fails we fail the whole deposit
 - Check account has enough balance
 	If fails we fail the whole deposit
 -  Prepare the withdrawal from the users account
	  If fails we fail the whole deposit
 - Call out to unit to originate ACH
	  If fails we need to call out to void the pending withdrawal and then mark deposit as failed
 - Commit the withdrawal
      If fails we fail the whole deposit. Requires manual intervention.
 - Set deposit as complete
*/

func DepositWorkflow(ctx workflow.Context, id string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Begin deposit")
	var depositActivity *Activity
	var mxActivity *_mx.Activity

	var deposit *Deposit
	err := workflow.ExecuteActivity(ctx, depositActivity.GetDeposit, id).Get(ctx, &deposit)
	if err != nil {
		logger.Error("error getting deposit", err)
		return err
	}

	var mxAcc *_mx.Account
	err = workflow.ExecuteActivity(ctx, mxActivity.GetMxAccountByFundingsource, deposit.FundingSourceId).Get(ctx, &mxAcc)
	if err != nil {
		logger.Error("error getting mx account.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, depositActivity.SetDepositState, id, Processing).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		mxActivity.StartBalanceAggregation,
		mxAcc.Guid,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("error starting balance aggregation", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, mxActivity.WaitForAggregation, mxAcc.Guid, 5, 5).Get(ctx, nil)
	if err != nil {
		logger.Error("error waiting for balance aggregation to complete", err)
		return err
	}

	var balance *_mx.AccountBalance
	err = workflow.ExecuteActivity(ctx, mxActivity.GetMxAccountBalance, mxAcc.Guid).Get(ctx, &balance)
	if err != nil {
		logger.Error("error getting account balance", err)
		return err
	}
	// Supporting only USD for now.
	if balance.AssetCode != "USD" && balance.AssetScale != 2 {
		logger.Error(fmt.Sprintf("Mx account is not in USD. assetCode=%s, assetScale=%d", balance.AssetCode, balance.AssetScale))
		return temporal.NewApplicationError("Mx account is not in USD", "ErrInternal")
	}
	if balance.Value <= 0 || uint64(balance.Value) < deposit.Amount {
		logger.Error(fmt.Sprintf("Insuffient balance. accountBalance=%+v, depositAmount=%d", balance, deposit.Amount))
		return temporal.NewApplicationError("Insuffient balance.", "ErrInternal")
	}

	// Prepare the withdrawal from the account (if insufficient mark as failed)
	var trxId string
	err = workflow.ExecuteActivity(ctx, depositActivity.CreatePendingDeposit, id).Get(ctx, &trxId)
	if err != nil {
		logger.Error("error creating pending transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, depositActivity.ProcessNoopDeposit, id).Get(ctx, nil)
	if err != nil {
		logger.Error("error processing noop transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, depositActivity.PostPendingDeposit, trxId).Get(ctx, nil)
	if err != nil {
		logger.Error("error posting pending transaction", err)
		return err
	}
	err = workflow.ExecuteActivity(ctx, depositActivity.SetDepositState, id, Complete).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}
	logger.Info("Workflow complete")

	return nil
}
