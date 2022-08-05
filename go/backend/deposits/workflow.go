package deposits

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/mx"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/*
 - Start balance aggregation
    If fails we fail the whole deposit
 - Check account has enough balance
 	If fails we fail the whole deposit
 - Call out to unit to originate ACH
	  If fails we need to call out to void the pending withdrawal and then mark deposit as failed
 - Make the funds available to the user's account
      If fails we fail the whole deposit. Requires manual intervention.
 - Set deposit as complete
*/

func DepositWorkflow(ctx workflow.Context, id string) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 35 * time.Second, // retry up to 3 times
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Begin deposit")
	var depositActivity *Activity
	var mxActivity *_mx.Activity
	var unitActivity *unit.Activity

	// Channels that webhooks will use to communicate with the workflow through.
	achStatusChannel := workflow.GetSignalChannel(ctx, "unit-user-ach-deposit")

	defer func() {
		if err != nil {
			cleanupCtx, _ := workflow.NewDisconnectedContext(ctx)
			_ = workflow.ExecuteActivity(cleanupCtx, depositActivity.SetDepositState, id, Failed).Get(cleanupCtx, nil)
		}
	}()

	var deposit *Deposit
	err = workflow.ExecuteActivity(ctx, depositActivity.GetDeposit, id).Get(ctx, &deposit)
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

	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(
			ctx,
			workflow.ActivityOptions{
				StartToCloseTimeout:    70 * time.Second, // we retry for up to 60s when waiting for aggregation
				ScheduleToCloseTimeout: 5 * time.Minute,  // to accomodate failures on waiting for aggregation
			},
		),
		mxActivity.WaitForAggregation,
		&mx.WaitForAggregationArgs{
			MxAccountGuid: mxAcc.Guid,
			MaxRetries:    5,
			PollInterval:  12 * time.Second,
		},
	).Get(ctx, nil)
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
	if env.IsProd() && (balance.AssetCode != "USD" || balance.AssetScale != 2) {
		err = fmt.Errorf("%w Mx account is not in USD. assetCode=%s, assetScale=%d", ErrInternal, balance.AssetCode, balance.AssetScale)
		logger.Error(err.Error())
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}
	if balance.Value <= 0 || uint64(balance.Value) < deposit.Amount {
		err = fmt.Errorf("%w Insufficient funding source balance. accountBalance=%+v, depositAmount=%d", ErrInternal, balance, deposit.Amount)
		logger.Error(err.Error())
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}

	err = workflow.ExecuteActivity(
		ctx,
		unitActivity.UnitInitiateUserDeposit,
		&unit.InitiateUserDepositArgs{
			DepositID:       deposit.ID,
			AccountID:       deposit.AccountID,
			FundingsourceID: deposit.FundingSourceId,
			Amount:          deposit.Amount,
			Description:     "Fynbos", // this will show up on the statement of unit counterparty
		},
	).Get(ctx, nil)
	if err != nil {
		logger.Error("error initiating user deposit on unit", err)
		return err
	}

	// we don't send the entire achPayment as it may contain sensitive info such as
	// account routing numbers.
	var achEventType string
	for {
		if unit.IsAchComplete(unit.EventType(achEventType)) {
			// We don't need to wait for webhooks if the ach has been completed un/successfully.
			break
		}

		achStatusChannel.Receive(ctx, &achEventType)
	}

	if !unit.IsAchSuccessful(unit.EventType(achEventType)) {
		err = fmt.Errorf("%w ACH failed. achStatus=%s", ErrInternal, achEventType)
		logger.Error(err.Error())
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
	}

	// TODO: ledger transfers and account transactions should be separated.
	err = workflow.ExecuteActivity(ctx, depositActivity.CreateAchDepositTransactions, deposit.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("error creating account deposit transaction", err)
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
