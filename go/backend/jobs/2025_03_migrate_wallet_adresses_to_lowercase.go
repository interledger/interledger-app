package jobs

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type wa struct {
	ID string
}

func MigrateWalletAddressesToLowercaseJob(ctx workflow.Context) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var wallets []wa
	err := workflow.ExecuteActivity(ctx, a.ListAllWalletAddresses).Get(ctx, &wallets)
	if err != nil {
		return nil, err
	}

	var failedWallets []string
	for _, w := range wallets {
		err := workflow.ExecuteActivity(ctx, a.WalletAddressToLowercase, w.ID).Get(ctx, nil)
		if err != nil {
			failedWallets = append(failedWallets, w.ID)
		}
	}

	return failedWallets, nil
}

func (a *Activity) ListAllWalletAddresses(ctx context.Context) ([]wa, error) {
	var waIDs []wa
	err := a.b.DB().SelectContext(ctx, &waIDs, "SELECT id FROM wallet_addresses")
	if err != nil {
		return nil, err
	}
	return waIDs, nil
}

func (a *Activity) WalletAddressToLowercase(ctx context.Context, walletID string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE wallet_addresses SET url=lower(url) WHERE id = $1", walletID)
	return err
}
