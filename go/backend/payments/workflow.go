package payments

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

func OutgoingPaymentWorkflow(ctx workflow.Context, id string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Begin outgoing payment")

	var a *Activity

	err := workflow.ExecuteActivity(ctx, a.SetOutgoingPaymentState, id, Processing).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}

	// Generate the outgoing payment transferID
	var transferID string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.NewString()
	}).Get(&transferID)
	if err != nil {
		logger.Error("error getting payments transfer ID side effect", err)
		return err
	}

	var trxId string
	err = workflow.ExecuteActivity(ctx, a.CreatePendingOutgoingPayment, id, transferID).Get(ctx, &trxId)
	if err != nil {
		logger.Error("error creating pending transaction", err)
		return err
	}

	defer func() {

		// Handle non-retryable errors by voiding the outgoing payment and releasing the liquidity.
		if !errors.Is(ctx.Err(), workflow.ErrCanceled) {
			return
		}

		// When the Workflow is canceled, it has to get a new disconnected context to execute any Activities
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		wfErr := workflow.ExecuteActivity(newCtx, a.VoidPendingOutgoingPayment, trxId).Get(ctx, nil)
		if wfErr != nil {
			logger.Error("VoidPendingOutgoingPayment cleanup failed", "Error", wfErr)
		}
	}()

	err = workflow.ExecuteActivity(ctx, a.ProcessNoopOutgoingPayment, id).Get(ctx, nil)
	if err != nil {
		logger.Error("error processing noop transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.PostPendingOutgoingPayment, trxId).Get(ctx, nil)
	if err != nil {
		logger.Error("error posting pending transaction", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetOutgoingPaymentState, id, Complete).Get(ctx, nil)
	if err != nil {
		logger.Error("error setting state", err)
		return err
	}

	logger.Info("Workflow complete")

	return nil
}
