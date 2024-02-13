package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_ops "gitlab.com/fynbos/backend/providers/pti/ops"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RerunCreatePTIWallets(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var walletIDs []string
	err := workflow.ExecuteActivity(ctx, a.ListWalletsWithPTIUsers).Get(ctx, &walletIDs)
	if err != nil {
		return err
	}

	for _, w := range walletIDs {
		_ = workflow.ExecuteChildWorkflow(ctx, pti_ops.CreateWalletWorkflow, pti.CreateWalletArgs{
			WalletID: w,
			Currency: currency.USD,
			Nickname: "USD Balance",
			Title:    "USD Balance",
		})
	}

	return nil
}

func (a *Activity) ListWalletsWithPTIUsers(ctx context.Context) ([]string, error) {
	var walletIDs []string
	_, err := a.b.DB().ExecContext(ctx, "SELECT wallet_id FROM 'pti_users';")
	if err != nil {
		return nil, err
	}

	return walletIDs, nil
}
