package workflows

import (
	"time"

	"gitlab.com/fynbos/backend/providers/mx/activities"

	"go.temporal.io/sdk/workflow"
)

type MxCreateFundingsourceWorkflowArgs struct {
	MxAccountGuid string
	AccountID     string
	Name          string
}

func MxCreateFundingsourceWorkflow(ctx workflow.Context, args *MxCreateFundingsourceWorkflowArgs) error {
	var a *activities.Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 35 * time.Second, // retry up to 3 times
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Creating fundingsource from mx bank account.")

	err := workflow.ExecuteActivity(ctx, a.CreateUnitCounterParty, args.MxAccountGuid).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit counter party.")
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.CreateFundingSource,
		args.MxAccountGuid,
		args.Name,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create funding source.")
		return err
	}

	return nil
}
