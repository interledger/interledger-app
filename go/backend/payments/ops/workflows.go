package ops

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/payments"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentWorkflow(ctx workflow.Context, id string) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute, // Tabapay calls may take up to 90s
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			// NonRetryableErrorTypes: []string{},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing payment", "id", id)

	// OFAC and compliance checks
	err := workflow.ExecuteChildWorkflow(ctx, gmt_workflows.GMTComplianceChecksWorkflow, id).Get(ctx, nil)
	if err != nil {
		logger.Error("OFAC and compliance checks failed. paymentID=", id, "err", err)
		innerErr := workflow.ExecuteActivity(ctx, a.SendPaymentFailedEmail, id).Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to send payment failed email", "paymentID=", id, "err", innerErr)
		}
		return
	}

	// launch payout and payin workflows in parallel
	selector := workflow.NewSelector(ctx)
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Namespace:         "payments",
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})
	payinFuture := workflow.ExecuteChildWorkflow(childCtx, PayinWorkflow, id)
	selector.AddFuture(payinFuture, func(f workflow.Future) {
		innerErr := f.Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Payin worfklow failed. paymentID=", id, "err", innerErr)
		}
		err = innerErr
	})

	payoutFuture := workflow.ExecuteChildWorkflow(childCtx, PayoutWorkflow, id)
	selector.AddFuture(payoutFuture, func(f workflow.Future) {
		innerErr := f.Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Payout worfklow failed. paymentID=", id, "err", innerErr)
			// rollback payin
		}
		err = innerErr
	})

	for {
		if !selector.HasPending() {
			break
		}

		selector.Select(ctx)
		if err != nil {
			innerErr := workflow.ExecuteActivity(ctx, a.SendPaymentFailedEmail, id).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to send payment failed email", "paymentID=", id, "err", innerErr)
				return
			}
		}
	}

	// run the next steps in parallel and log any errors.
	completedSelector := workflow.NewSelector(ctx)
	setCompleteFuture := workflow.ExecuteActivity(ctx, a.SetPaymentState, id, payments.StateCompleted)
	completedSelector.AddFuture(setCompleteFuture, func(f workflow.Future) {
		innerErr := f.Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to set state payment to complete. paymentID=", id, "err", innerErr)
			// TODO: notify someone
		}
	})

	sendPaymentReceivedFuture := workflow.ExecuteActivity(ctx, a.SendPaymentReceivedEmail, id)
	completedSelector.AddFuture(sendPaymentReceivedFuture, func(f workflow.Future) {
		innerErr := f.Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to send payment received email", "paymentID=", id, "err", innerErr)
			// TODO: notify someone
		}
	})

	sendPaymentSentFuture := workflow.ExecuteActivity(ctx, a.SendPaymentSentEmail, id)
	completedSelector.AddFuture(sendPaymentSentFuture, func(f workflow.Future) {
		innerErr := f.Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to send payment sent email", "paymentID=", id, "err", innerErr)
			// TODO: notify someone
		}
	})

	for {
		if !completedSelector.HasPending() {
			break
		}

		completedSelector.Select(ctx)
	}

	logger.Info("Payment processed. paymentID=", id)
}

func PayinWorkflow(ctx workflow.Context, paymentID string) error {
	return nil
}

func PayoutWorkflow(ctx context.Context, paymentID string) error {
	return nil

}
