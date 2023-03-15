package ops

import (
	"time"

	"gitlab.com/fynbos/backend/providers/gmt"
	"go.temporal.io/sdk/workflow"
)

func OnboardUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("OnboardUserWorkflow workflow started", "walletID", walletID)

	err := workflow.ExecuteActivity(ctx, a.CheckOFAC, walletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do OFAC checks", "err", err)
	}

	err = workflow.ExecuteActivity(ctx, a.IndividualCompliance, walletID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
	}

	return "TODO", nil
}

func ACH2ACHTransferWorkflow(ctx workflow.Context, args gmt.TransfersArgs) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ACH2ACHTransferWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	err := workflow.ExecuteActivity(ctx, a.CheckOFAC, args).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do OFAC checks", "err", err)
	}

	err = workflow.ExecuteActivity(ctx, a.ACHCompliance, args).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to do compliance checks", "err", err)
	}

	return "TODO", nil
}
