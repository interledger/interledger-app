package ops

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func StartTravelRuleEmailSending(b ActivityBackends) {
	// Every day at midnight UTC
	schedule := "0 0 * * *"
	workflowID := "cron_xago_travel_rule_email"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          schedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, TravelRuleEmailWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func TravelRuleEmailWorkflow(ctx workflow.Context) error {
	var a *Activity
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	})

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting xago travel rule email")

	var records []TravelRuleRecord
	if err := workflow.ExecuteActivity(ctx, a.GetTravelRuleRecords).Get(ctx, &records); err != nil {
		return err
	}

	if len(records) == 0 {
		logger.Info("No travel rule payments to report")
		return nil
	}

	var csvBytes []byte
	if err := workflow.ExecuteActivity(ctx, a.BuildTravelRuleCSV, records).Get(ctx, &csvBytes); err != nil {
		return err
	}

	if err := workflow.ExecuteActivity(ctx, a.SendTravelRuleEmail, csvBytes).Get(ctx, nil); err != nil {
		return err
	}

	return workflow.ExecuteActivity(ctx, a.MarkTravelRuleRecordsAsReported, records).Get(ctx, nil)
}

func StartDepositsPolling(b ActivityBackends) {
	schedule := "0 */1 * * *"
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_xago_deposits_poll"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          schedule,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, XagoDepositPollWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func XagoDepositPollWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	logger.Info("starting xago deposits")

	var deposits []external.Deposit
	err := workflow.ExecuteActivity(ctx, a.PollDeposits).Get(ctx, &deposits)
	if err != nil {
		return err
	}
	// No new deposits, so nothing to do
	if len(deposits) == 0 {
		return nil
	}

	logger.Info("Adding new deposits", "len", len(deposits))

	err = workflow.ExecuteActivity(ctx, a.CreateDepositTransactions, deposits).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveDeposits, deposits).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
