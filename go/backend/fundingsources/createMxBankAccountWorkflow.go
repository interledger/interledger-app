package fundingsources

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type CreateMxBankAccountWorkflowArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required"`
	MxUserGuid      string `validate:"required"`
	MxMemberGuid    string `validate:"required"`
}

func CreateMxBankAccountWorkflow(ctx workflow.Context, args *CreateMxBankAccountWorkflowArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Creating mx bank account")

	mxAccountID := ""
	err := workflow.ExecuteActivity(ctx, a.GetSelectedMxAccountGuid, args.MxUserGuid, args.MxMemberGuid).Get(ctx, &mxAccountID)
	if err != nil {
		logger.Error("Failed to find mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.CreateMxAccount,
		args.FundingSourceID,
		args.MxUserGuid,
		args.MxMemberGuid,
		mxAccountID,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StartIdentityAggregation, args.FundingSourceID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to start mx identity aggregation.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.WaitForIdentityAggregation, args.FundingSourceID).Get(ctx, nil)
	if err != nil {
		logger.Error("Mx identity aggregation failed.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyOwnership, args.FundingSourceID, args.IdentityID).Get(ctx, nil)
	if err != nil {
		logger.Error("Bank account is not owned by user.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetMask, args.FundingSourceID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set mask for funding source.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUnitCounterParty, args.FundingSourceID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit counter party.")
		return err
	}

	return nil
}
