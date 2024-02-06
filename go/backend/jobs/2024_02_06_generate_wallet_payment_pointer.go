package jobs

import (
	"context"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"time"

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

	err := workflow.ExecuteActivity(ctx, a.UpdateTransactionsIncomingToReceived).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionsOutgoingToSent).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) GenerateWalletPaymentPointers(ctx context.Context) error {
	var ws []wallets.Wallet
	err := a.b.DB().SelectContext(
		ctx,
		ws,
		"select * from wallets where id not in (select wallet_id from rafiki_payment_pointers);",
	)
	if err != nil {
		return err
	}

	for _, w := range ws {
		// create rafiki payment pointer
		err = a.b.Rafiki().CreatePaymentPointer(ctx, w, "USD")
		if err != nil {
			log.Error("couldn't create payment point for wallet", zap.String("walletAddress", w.AddressString()), zap.String("walletID", w.ID))
		}
	}

	return nil
}
