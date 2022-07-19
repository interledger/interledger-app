package flows

import (
	"time"

	"gitlab.com/fynbos/backend/deposits"

	"gitlab.com/fynbos/backend/deposits/ops"
	"go.temporal.io/sdk/workflow"
)

/*
 1. Prepare the withdrawal from the users account
	If fails we fail the whole deposit
 2. Call out to provider to initiate (potentially async)
	If fails we need to call out to void the pending withdrawal and then mark deposit as failed
 3. Commit the withdrawal
 4. Set deposit as complete
*/

func DepositWorkflow(ctx workflow.Context, id string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Begin deposit")

	var a *ops.Activity
	err := workflow.ExecuteActivity(ctx, a.SetDepositState, id, deposits.Processing).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}

	// Prepare the withdrawal from the account (if insufficient mark as failed)
	var trxId string
	err = workflow.ExecuteActivity(ctx, a.CreatePendingDeposit, id).Get(ctx, &trxId)
	if err != nil {
		logger.Error("error creating pending transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.ProcessNoopDeposit, id).Get(ctx, nil)
	if err != nil {
		logger.Error("error processing noop transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.PostPendingDeposit, trxId).Get(ctx, nil)
	if err != nil {
		logger.Error("error posting pending transaction", err)
		return err
	}
	err = workflow.ExecuteActivity(ctx, a.SetDepositState, id, deposits.Complete).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}
	logger.Info("Workflow complete")

	return nil
}
