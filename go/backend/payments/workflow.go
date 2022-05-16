package payments

import (
	"go.temporal.io/sdk/workflow"
	"time"
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

	var trxId string
	err = workflow.ExecuteActivity(ctx, a.CreatePendingOutgoingPayment, id).Get(ctx, &trxId)
	if err != nil {
		logger.Error("error creating pending transaction", err)
		return err
	}

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
