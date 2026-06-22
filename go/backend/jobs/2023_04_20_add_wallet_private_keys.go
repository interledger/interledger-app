package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/db"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func AddWalletPrivateKeysWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("AddWalletPrivateKeysWorkflow workflow started")

	// Provision private keys for all wallets
	err := workflow.ExecuteActivity(ctx, a.RegisterWalletsPrivateKeyActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("RegisterWalletsPrivateKeyActivity Activity failed.", "Error", err)
		return err
	}

	return nil
}

func (a *Activity) RegisterWalletsPrivateKeyActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	wallets, err := a.b.Wallets().ListAll(ctx, db.Pagination{
		PageToken: "",
		PageSize:  50,
	})
	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		logger.Info("Registering private key for wallet", "wallet", wallet.ID)
		err := a.b.Keys().ProvisionPrivateKey(ctx, wallet.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
