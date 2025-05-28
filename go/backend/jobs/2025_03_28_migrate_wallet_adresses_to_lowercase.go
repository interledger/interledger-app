package jobs

import (
	"context"
	"os"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type wa struct {
	ID string
}

func MigrateWalletAddressesToLowercaseJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.TransformWalletAddressesToLowerCase).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) TransformWalletAddressesToLowerCase(ctx context.Context) error {
	connString := os.Getenv("RAFIKI_DB_URL")
	rafikiDb, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer rafikiDb.Close()

	tx1, err := a.b.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	tx2, err := rafikiDb.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		err = tx1.Rollback()
		err = tx2.Rollback()
	}()

	_, err = tx1.ExecContext(ctx, "UPDATE wallet_addresses SET url = LOWER(url)")
	if err != nil {
		return err
	}

	_, err = tx2.ExecContext(ctx, `UPDATE "walletAddresses" SET url = LOWER(url)`)
	if err != nil {
		return err
	}

	err = tx1.Commit()
	if err != nil {
		return err
	}

	err = tx2.Commit()
	if err != nil {
		return err
	}

	return nil
}
