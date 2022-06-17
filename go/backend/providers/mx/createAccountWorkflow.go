package mx

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type CreateMxAccountWorkflowArgs struct {
	ID                string `validate:"uuid4"`
	UserGuid          string `validate:"required"`
	MemberGuid        string `validate:"required"`
	AccountID         string `validate:"required"`
	IdentityID        string `validate:"uuid4"`
	FundingsourceName string `validate:"required"`
}

// The output of this workflow is a fundingsource identified by args.ID.
// Mx is used to pull the account owner information and verify ownership. The account routing
// numbers are then pulled to create a unit counterparty which is also mapped to the fundingsource.
func CreateMxAccountWorkflow(ctx workflow.Context, args *CreateMxAccountWorkflowArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    10 * time.Second,
		ScheduleToCloseTimeout: 15 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("Creating mx bank account")

	// we tie the funding source to this workflow by setting the fundingsourceID to the supplied ID.
	fundingsourceID := args.ID
	mxAccountGuid := "" // the id that mx generates
	err := workflow.ExecuteActivity(ctx, a.GetSelectedMxAccountGuid, args.UserGuid, args.MemberGuid).Get(ctx, &mxAccountGuid)
	if err != nil {
		logger.Error("Failed to find mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.CreateMxAccount,
		fundingsourceID,
		args.AccountID,
		args.UserGuid,
		args.MemberGuid,
		mxAccountGuid,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create mx account.", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StartIdentityAggregation, mxAccountGuid).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to start mx identity aggregation.", err)
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.WaitForIdentityAggregation,
		mxAccountGuid,
		10,
		time.Second,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Mx identity aggregation failed.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.VerifyOwnership, mxAccountGuid).Get(ctx, nil)
	if err != nil {
		logger.Error("Bank account is not owned by user.")
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUnitCounterParty, mxAccountGuid).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create unit counter party.")
		return err
	}

	err = workflow.ExecuteActivity(
		ctx,
		a.CreateFundingSource,
		mxAccountGuid,
		args.FundingsourceName,
	).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create funding source.")
		return err
	}

	return nil
}
