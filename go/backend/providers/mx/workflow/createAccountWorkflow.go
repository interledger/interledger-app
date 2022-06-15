package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type CreateMxAccountWorkflowArgs struct {
	ID           string `validate:"uuid4"`
	AccountID    string `validate:"required"`
	MxUserGuid   string `validate:"required"`
	MxMemberGuid string `validate:"required"`
}

func CreateMxAccountWorkflow(ctx workflow.Context, args *CreateMxAccountWorkflowArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Creating mx bank account")

	mxAccountGuid := "" // the id that mx generates
	err := workflow.ExecuteActivity(ctx, a.GetSelectedMxAccountGuid, args.MxUserGuid, args.MxMemberGuid).Get(ctx, &mxAccountGuid)
	if err != nil {
		logger.Error("Failed to find mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.CreateMxAccount,
		args.ID,
		args.AccountID,
		args.MxUserGuid,
		args.MxMemberGuid,
		mxAccountGuid,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StartIdentityAggregation, args.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to start mx identity aggregation.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.WaitForIdentityAggregation, args.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("Mx identity aggregation failed.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyOwnership, args.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("Bank account is not owned by user.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUnitCounterParty, args.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit counter party.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateFundingSource, args.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create funding source.")
		return err
	}

	return nil
}
