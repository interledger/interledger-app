package workflows

import (
	"gitlab.com/fynbos/backend/kyc"
	"go.temporal.io/sdk/workflow"
	"time"
)

type StartKYCArgs struct {
	WalletID string
}

func StartKYC(ctx workflow.Context, args StartKYCArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.SetKYCStatus, args.WalletID, kyc.StatusApproved).Get(ctx, nil)

	return err
}
