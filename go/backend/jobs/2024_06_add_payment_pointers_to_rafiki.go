package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateRafikiPaymentPointersJob(ctx workflow.Context) ([]string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var wallets []wallets.Wallet
	err := workflow.ExecuteActivity(ctx, a.ListAllWallets).Get(ctx, &wallets)
	if err != nil {
		return nil, err
	}

	var failedWallets []string
	for _, w := range wallets {
		err := workflow.ExecuteActivity(ctx, a.AddWalletToRafiki, w.ID).Get(ctx, nil)
		if err != nil {
			failedWallets = append(failedWallets, w.ID)
		}
	}

	return failedWallets, nil
}

func (a *Activity) ListAllWallets(ctx context.Context) ([]wallets.Wallet, error) {
	return a.b.Wallets().ListAll(ctx, db.Pagination{})
}

func (a *Activity) AddWalletToRafiki(ctx context.Context, walletID string) error {
	w, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, wallets.ErrNoWalletFound) {
		return temporal.NewNonRetryableApplicationError("create rafiki payment pointer job: no wallet found", "ErrInternal", err)
	}
	if err != nil {
		return err
	}
	_, err = a.b.Rafiki().CreatePaymentPointer(ctx, *w)
	if err != nil {
		return err
	}

	return nil
}
