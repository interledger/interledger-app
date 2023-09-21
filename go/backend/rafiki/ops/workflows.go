package ops

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func PayoutIncomingPaymentsWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var payouts []Payout
	err := workflow.ExecuteActivity(ctx, a.ListPayouts).Get(ctx, &payouts)
	if err != nil {
		logger.Error("failed to list incoming payments set for payout", "err", err)
		return err
	}

	for _, p := range payouts {
		var paymentID string
		err = workflow.ExecuteActivity(ctx, a.CreatePayoutPayment, p).Get(ctx, &paymentID)
		if err != nil {
			logger.Error("failed to create payment for payout", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.ConfirmPayment, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to create payment for payout", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.AddPaymentRef, p.ID, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to add payment ref to incoming payment payout", "err", err)
			return err
		}
	}

	return nil
}
