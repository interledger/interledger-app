package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RemoveCustodialKeysJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job RemoveCustodialKeys: removing custodial keys for all wallets")

	err := workflow.ExecuteActivity(ctx, a.RemoveCustodialKeys).Get(ctx, nil)
	if err != nil {
		return err
	}

	log.Info("Completed job RemoveCustodialKeys")
	return nil
}

func (a *Activity) RemoveCustodialKeys(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	wallets, err := a.b.Wallets().ListAll(ctx, db.Pagination{
		PageToken: "",
		PageSize:  50000,
	})

	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		logger.Info("removing custodial keys for wallet", "wallet", wallet.ID)
		err = a.b.Keys().RemoveCustodialKeysForWallet(ctx, wallet.ID)
		if err != nil {
			return err
		}
	}

	log.Info("Deleted custodial wallet keys")
	return nil
}
