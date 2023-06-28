package jobs

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"

	"gitlab.com/fynbos/backend/wallets"

	"go.temporal.io/sdk/workflow"
)

func MigratePaymentPointers(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("MigratePaymentPointers workflow started")

	return workflow.ExecuteActivity(ctx, a.MigratePaymentPointers).Get(ctx, nil)
}

func (a *Activity) MigratePaymentPointers(ctx context.Context) error {
	type PP struct {
		WalletID string `db:"wallet_id"`
		URL      string `db:"url"`
	}

	var AllPP []PP
	err := a.b.DB().SelectContext(ctx, &AllPP, "SELECT wallet_id, url FROM payment_pointers")
	if err != nil {
		return err
	}

	var createdCnt, skipCnt int
	for _, pp := range AllPP {
		_, err = a.b.Wallets().AddAddress(ctx, pp.WalletID, pp.URL)
		if errors.Is(err, wallets.ErrDuplicateWallet) {
			skipCnt++
			continue
		}
		if err != nil {
			return err
		}
		createdCnt++
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Payment Pointer Migration", "payment_pointer_cnt", len(AllPP), "wallets_created", createdCnt, "wallets_already_exits", skipCnt)

	return nil
}
