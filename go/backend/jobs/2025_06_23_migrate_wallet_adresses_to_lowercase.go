package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

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
	connString := a.cfg.RafikiDBURL
	rafikiDB, err := DbConnection(connString)
	if err != nil {
		log.Error("Error establishing db connection: %v", zap.Error(err))
		return err
	}
	defer rafikiDB.Close()

	tx1, err := a.b.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	tx2, err := rafikiDB.BeginTxx(ctx, nil)
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
