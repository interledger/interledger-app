package ops

import (
	"fmt"
	"time"

	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/providers/tabapay"
	temporal_utils "gitlab.com/fynbos/backend/temporal/utils"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func PaymentWorkflow(ctx workflow.Context, id string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{"ErrInvalidStateTransition", "ErrNotFound"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing payment", "id", id)
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Namespace:         "payments",
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})

	err := workflow.ExecuteActivity(ctx, a.SetPaymentStateProcessing, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set payment state to processing. paymentID=", id, "err", err)
		return err
	}

	// OFAC and compliance checks
	err = workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTComplianceChecksWorkflow, id).Get(childCtx, nil)
	if err != nil {
		logger.Error("GMT compliance failed for payment", "payment_id", id, "error", err)
		innerErr := workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
		if innerErr != nil {
			logger.Error("Failed to set payment status to failed. paymentID=", id, "err", innerErr)
			return innerErr
		}

		return nil
	}

	// launch payout and payin workflows in parallel
	selector := workflow.NewSelector(ctx)

	childPayInCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            fmt.Sprintf(payinWorkflowFmt, id),
		Namespace:             "payments",
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_TERMINATE,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	})
	payinFuture := workflow.ExecuteChildWorkflow(childPayInCtx, PayinWorkflow, id)
	selector.AddFuture(payinFuture, func(f workflow.Future) {
		innerErr := f.Get(childPayInCtx, nil)
		if innerErr != nil {
			logger.Error("Payin worfklow failed. paymentID=", id, "err", innerErr)
		}
		err = innerErr
	})

	childPayoutCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            fmt.Sprintf(payoutWorkflowFmt, id),
		Namespace:             "payments",
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_TERMINATE,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	})
	payoutFuture := workflow.ExecuteChildWorkflow(childPayoutCtx, PayoutWorkflow, id)
	selector.AddFuture(payoutFuture, func(f workflow.Future) {
		innerErr := f.Get(childPayoutCtx, nil)
		if innerErr != nil {
			logger.Error("Payout worfklow failed. paymentID=", id, "err", innerErr)
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

	var success bool
	err = workflow.ExecuteActivity(ctx, a.CheckPaymentSuccess, id).Get(ctx, &success)
	if err != nil {
		return err
	}

	if !success {
		// Mark payment as a failure and exit workflow.
		err = workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetPaymentStateComplete, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set state payment to complete. paymentID=", id, "err", err)
		return err
	}

	var we workflow.Execution
	err = workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTNotifyCompleted, id).GetChildWorkflowExecution().Get(childCtx, &we)
	// Child workflow execution has started. We can return and GMT will carry-on on its own
	if err != nil {
		logger.Error("Failed to notify GMT of completed payment", "err", err)
	}

	return nil
}

func PayinWorkflow(ctx workflow.Context, paymentID string) error {
	var a *Activity

	accountsAO := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Minute,
			BackoffCoefficient: 5, // Retry in 1 minute, then 5, then 25. Give tabapay a chance
		},
	}
	accountsCtx := workflow.WithActivityOptions(ctx, accountsAO)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	// Create the outgoing transaction
	var txID string
	err := workflow.ExecuteActivity(ctx, a.CreatePayInTransaction, paymentID).Get(ctx, &txID)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetSendTransactionID, paymentID, txID).Get(ctx, nil)
	if err != nil {
		return err
	}

	// TODO: decouple this from tabapay.
	var externalRef string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&externalRef)
	if err != nil {
		return err
	}

	var accountTX tabapay.Transaction
	err = workflow.ExecuteActivity(accountsCtx, a.PullFromAccount, paymentID, externalRef).Get(accountsCtx, &accountTX)
	if temporal_utils.IsNonRetryableError(err) || temporal_utils.IsMaxRetryError(err) {
		// Signal the Pay Out workflow that the payout has failed, it will know what to do.
		innerErr := workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:    paymentID,
			PayInSuccess: false,
		}).Get(ctx, nil)
		if innerErr != nil {
			return innerErr
		}
	}
	if err != nil {
		return err
	}

	if tabapay.IsTransactionStatusUnknown(accountTX) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again. id=", accountTX.ID)
		err = workflow.ExecuteActivity(accountsCtx, a.GetCardTransaction, accountTX.ID).Get(accountsCtx, &accountTX)
		if err != nil {
			logger.Error("failed to get tabapay send transaction", "err", err)
			// Notify the Payout worklfow of failure
			innerErr := workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
				PaymentID:    paymentID,
				PayInSuccess: false,
			}).Get(ctx, nil)
			if innerErr != nil {
				return innerErr
			}

			return err
		}
	}

	// Notify payout tx of success or failure and it can get on with it
	err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
		PaymentID:    paymentID,
		PayInSuccess: tabapay.IsSuccessfulTransaction(accountTX),
	}).Get(ctx, nil)
	if err != nil {
		return err
	}

	if !tabapay.IsSuccessfulTransaction(accountTX) {
		// Mark transaction as a failure and stop the workflow
		return workflow.ExecuteActivity(ctx, a.UpdateTransactionState, txID, transactions.StateFailed).Get(ctx, nil)
	}

	err = workflow.ExecuteActivity(accountsCtx, a.AddPayInTransfer, paymentID, accountTX.ID).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Now we wait for the payout to complete
	var signal PaySignal
	signalChan := workflow.GetSignalChannel(ctx, signalChanName)
	for {
		signalChan.Receive(ctx, &signal)
		if signal.PaymentID != paymentID {
			logger.Warn("Received signal for payment wrong payment", "payment_id", paymentID, "received", signal.PaymentID)
			continue
		}
		break
	}
	if signal.PayOutSuccess {
		// Mark transaction as a success
		return workflow.ExecuteActivity(ctx, a.UpdateTransactionState, txID, transactions.StateCompleted).Get(ctx, nil)
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, txID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}
	// Payout was unsuccessful. Rollback the account pull
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Namespace:         "payments",
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})

	return workflow.ExecuteChildWorkflow(childCtx, RollbackPayInWorkflow, paymentID).Get(childCtx, nil)
}

type PaySignal struct {
	PaymentID     string
	PayInSuccess  bool
	PayOutSuccess bool
}

const (
	signalChanName    = "payment_signals"
	payinWorkflowFmt  = "payment_pay_in_%s"
	payoutWorkflowFmt = "payment_pay_out_%s"
)

func PayoutWorkflow(ctx workflow.Context, paymentID string) error {
	var a *Activity

	accountsAO := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Minute,
			BackoffCoefficient: 5, // Retry in 1 minute, then 5, then 25. Give tabapay a chance
		},
	}
	accountsCtx := workflow.WithActivityOptions(ctx, accountsAO)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var signal PaySignal
	signalChan := workflow.GetSignalChannel(ctx, signalChanName)
	for {
		signalChan.Receive(ctx, &signal)
		if signal.PaymentID != paymentID {
			logger.Warn("Received signal for payment wrong payment", "payment_id", paymentID, "received", signal.PaymentID)
			continue
		}

		if signal.PayInSuccess {
			break
		}

		// Funding was unsuccessful, nothing more to do
		return nil
	}

	// Funding was a success now, payout
	var externalRef string
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return tabapay.NewReferenceID()
	}).Get(&externalRef)
	if err != nil {
		return err
	}

	var accountTX tabapay.Transaction
	err = workflow.ExecuteActivity(accountsCtx, a.PushToAccount, paymentID, externalRef).Get(accountsCtx, &accountTX)
	if temporal_utils.IsNonRetryableError(err) || temporal_utils.IsMaxRetryError(err) {
		// Signal the Pay In workflow that the payout has failed, it will know what to do.
		err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:     paymentID,
			PayOutSuccess: false,
		}).Get(ctx, nil)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	if tabapay.IsTransactionStatusUnknown(accountTX) {
		// check again in 90 sec
		_ = workflow.Sleep(ctx, time.Second*90)
		logger.Info("Tabapay transaction status unknown. Checking again. id=", accountTX.ID)
		err = workflow.ExecuteActivity(accountsCtx, a.GetCardTransaction, accountTX.ID).Get(accountsCtx, &accountTX)
		if err != nil {
			logger.Error("failed to get tabapay receive transaction", "err", err)
			// Notify the Payout worklfow of failure
			innerErr := workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
				PaymentID:     paymentID,
				PayOutSuccess: false,
			}).Get(ctx, nil)
			if innerErr != nil {
				return innerErr
			}
			return err
		}
	}

	// Notify payin tx of success or failure and it can get on with it
	err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
		PaymentID:     paymentID,
		PayOutSuccess: tabapay.IsSuccessfulTransaction(accountTX),
	}).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Payout was unsuccessful, nothing more to do.
	if !tabapay.IsSuccessfulTransaction(accountTX) {
		return nil
	}

	// Create the incoming transaction
	var txID string
	err = workflow.ExecuteActivity(ctx, a.CreatePayoutTransaction, paymentID, accountTX.ID).Get(ctx, &txID)
	if err != nil {
		return err
	}

	// TODO: should we do 2-phase for the payout transaction?
	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, txID, transactions.StateCompleted).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.SetReceiveTransactionID, paymentID, txID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func RollbackPayInWorkflow(ctx workflow.Context, paymentID string) error {
	var a *Activity

	logger := workflow.GetLogger(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 95 * time.Second, // May take up to 90 seconds.
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    10, // Tabapay will block us if we keep retrying, but we will try at least 10 times, with huge breaks in between so we don't hit rate limits
			BackoffCoefficient: 2,
			InitialInterval:    time.Minute * 10,
			MaximumInterval:    time.Hour,
		},
	})

	err := workflow.ExecuteActivity(ctx, a.RollbackPullFromAccount, paymentID).Get(ctx, nil)
	if err != nil {
		if temporal_utils.IsNonRetryableError(err) || temporal_utils.IsMaxRetryError(err) {
			logger.Error("Final failure to rollback tabapay card transaction", "err", err, "payment_id", paymentID)
		}

		return err
	}

	// Insert rollback transfer
	err = workflow.ExecuteActivity(ctx, a.AddPayInRollbackTransfer, paymentID).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction transfer", "Error", err)
		return err
	}

	return nil
}
