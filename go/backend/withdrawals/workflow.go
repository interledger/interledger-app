package withdrawals

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

/*
 1. Prepare the withdrawal from the users account
	If fails we fail the whole withdrawal
 2. Call out to provider to initiate (potentially async)
	If fails we need to call out to void the pending withdrawal and then mark deposit as failed
 3. Commit the withdrawal
 4. Set withdrawal as complete
*/

func WithdrawalWorkflow(ctx workflow.Context, id string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Begin withdrawal")

	var a *Activity
	err := workflow.ExecuteActivity(ctx, a.SetWithdrawalState, id, Processing).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}

	var transferID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&transferID)
	if err != nil {
		logger.Error("error getting payments transfer ID side effect", err)
		return err
	}

	// Prepare the withdrawal from the account (if insufficient mark as failed)
	var trxId string
	err = workflow.ExecuteActivity(ctx, a.CreatePendingWithdrawalTransaction, id, transferID).Get(ctx, &trxId)
	if err != nil {
		logger.Error("error creating pending transaction", err)
		return err
	}

	defer func() {

		// Handle non-retryable errors by voiding the withdrawal and releasing the liquidity.
		if !errors.Is(ctx.Err(), workflow.ErrCanceled) {
			return
		}

		// When the Workflow is canceled, it has to get a new disconnected context to execute any Activities
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		wfErr := workflow.ExecuteActivity(newCtx, a.VoidPendingWithdrawalTransaction, trxId).Get(ctx, nil)
		if wfErr != nil {
			logger.Error("VoidPendingWithdrawalTransaction cleanup failed", "Error", wfErr)
		}
	}()

	err = workflow.ExecuteActivity(ctx, a.ProcessNoopWithdrawal, id).Get(ctx, nil)
	if err != nil {
		logger.Error("error processing noop transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.PostPendingWithdrawalTransaction, trxId).Get(ctx, nil)
	if err != nil {
		logger.Error("error posting pending transaction", err)
		return err
	}
	err = workflow.ExecuteActivity(ctx, a.SetWithdrawalState, id, Complete).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}
	logger.Info("Workflow complete")

	return nil
}
