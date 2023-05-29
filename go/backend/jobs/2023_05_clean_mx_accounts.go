package jobs

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func (a *Activity) ListMX(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	accs, err := a.b.LinkedAccounts().ListMXBankAccounts(ctx)
	if err != nil {
		return err
	}

	users, err := a.b.MX().ListUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		var found bool
		for _, acc := range accs {
			if user.WalletID == acc.WalletID {
				found = true
				break
			}
		}

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

	return nil
}
