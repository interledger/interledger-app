package ops

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func StartRafikiIncomingPaymentsPolling(b ActivityBackends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_rafiki_web_monetization_payouts_pti"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          "0 0 * * *",                                         // Every day at midnight
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, WebMonetizationPaymentsWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func WebMonetizationPaymentsWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var payouts []RafikiPayment
	err := workflow.ExecuteActivity(ctx, a.ListPaymentsToMake).Get(ctx, &payouts)
	if err != nil {
		logger.Error("failed to list outgoing payments set for payout", "err", err)
		return err
	}

	for _, p := range payouts {
		var paymentID string
		err = workflow.ExecuteActivity(ctx, a.CreateWebMonetizationPayment, p).Get(ctx, &paymentID)
		if err != nil {
			logger.Error("failed to create payment for web monetization", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.ConfirmPayment, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to confirm payment for web monetization", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.AddWebMonetizationPayment, p, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to add payment ref to incoming payment payout", "err", err)
			return err
		}
	}

	return nil
}
