package jobs

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const limit = 100
const batchSize = 100

func RemoveCustodialKeysJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	log.Info("Starting job RemoveCustodialkeys: removing custodial keys for all wallets")

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
		PageSize:  50,
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

	rows, err := a.b.DB().ExecContext(ctx, "DELETE FROM wallet_keys WHERE key_type = $1", keys.Custodial)
	if err != nil {
		return err
	}

	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	log.Info("Deleted custodial wallet keys", zap.Int64("rows-affected", affected))
	return nil
}
