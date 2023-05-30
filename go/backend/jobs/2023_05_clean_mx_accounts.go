package jobs

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func (a *Activity) CleanMX(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	accs, err := a.b.LinkedAccounts().ListMXBankAccounts(ctx)
	if err != nil {
		return err
	}

	walletIDs := make(map[string]bool)
	for _, acc := range accs {
		walletIDs[acc.WalletID] = true
	}

	users, err := a.b.MX().ListUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		_, found := walletIDs[user.WalletID]

		if !found {
			err = a.b.MX().DeleteExternalUser(ctx, user.GUID)
			if err != nil {
				return err
			}

			logger.Info("deleted unused MX account", "wallet_id", user.WalletID, "guid", user.GUID)
		}
	}

	return nil
}

func CleanMXAccounts(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.CleanMX).Get(ctx, nil)

	return err
}
