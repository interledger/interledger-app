package jobs

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func GenerateWalletPaymentPointersJob(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, a.GenerateWalletPaymentPointers).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) GenerateWalletPaymentPointers(ctx context.Context) error {
	var ids []string
	err := a.b.DB().SelectContext(
		ctx,
		&ids,
		"select id from wallets where id not in (select wallet_id from rafiki_payment_pointers);",
	)
	if err != nil {
		return err
	}

	for _, id := range ids {
		w, err := a.b.Wallets().Get(ctx, id)
		if err != nil {
			log.Error("could not get wallet", zap.String("walletID", id), zap.Error(err))
			continue
		}
		_, err = a.b.Rafiki().CreatePaymentPointer(ctx, *w)
		if err != nil {
			log.Error("couldn't create payment point for wallet", zap.String("walletAddress", w.AddressString()), zap.String("walletID", w.ID), zap.Error(err))
		}
	}

	return nil
}
