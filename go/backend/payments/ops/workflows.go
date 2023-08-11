package ops

import (
	"time"

	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentWorkflow(ctx workflow.Context, id string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
		RetryPolicy:         &temporal.RetryPolicy{
			// NonRetryableErrorTypes: []string{},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing payment", "id", id)
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Namespace:         "payments",
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})

	// OFAC and compliance checks
	err := workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTComplianceChecksWorkflow, id).Get(ctx, nil)
	if err != nil {
		innerErr := workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to set payment status to failed. paymentID=", id, "err", innerErr)
			return innerErr
		}

		return nil
	}

	err = workflow.ExecuteActivity(ctx, a.SetPaymentStateProcessing, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set payment state to processing. paymentID=", id, "err", err)
		return err
	}

	// launch payout and payin workflows in parallel
	selector := workflow.NewSelector(ctx)
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
		innerErr := f.Get(childCtx, nil)
		if innerErr != nil {
			logger.Error("Payout worfklow failed. paymentID=", id, "err", innerErr)
			// rollback payin
		}
		err = innerErr
	})

	for count := 0; count < 2; count++ {
		selector.Select(ctx)
		if err != nil {
			innerErr := workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to set payment status to failed. paymentID=", id, "err", innerErr)
				return innerErr
			}

			return nil
		}
	}

	err = workflow.ExecuteActivity(ctx, a.SetPaymentStateComplete, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set state payment to complete. paymentID=", id, "err", err)
	}

	var we workflow.Execution
	err = workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTNotifyCompleted, id).GetChildWorkflowExecution().Get(ctx, &we)
	// Child workflow execution has started. We can return and GMT will carry-on on its own
	if err != nil {
		logger.Error("Failed to notify GMT of completed payment", "err", err)
	}

	return nil
}

func PayinWorkflow(ctx workflow.Context, paymentID string) error {
	return nil
}

func PayoutWorkflow(ctx workflow.Context, paymentID string) error {
	return nil
}
