package ops

import (
	"fmt"
	"time"

	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	httplog "gitlab.com/fynbos/backend/providers/http"
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
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})

	err := workflow.ExecuteActivity(ctx, a.SetPaymentStateProcessing, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set payment state to processing. paymentID=", id, "err", err)
		return err
	}

	var receiverReady bool
	err = workflow.ExecuteActivity(ctx, a.CheckReceiverReady, id).Get(ctx, &receiverReady)
	if err != nil {
		return err
	}

	// OFAC and compliance checks, if the receiver is ready and signed up, otherwise the AwaitReceiverWorkflow will do it.
	if receiverReady {
		err = workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTComplianceChecksWorkflow, id).Get(childCtx, nil)
		if err != nil {
			logger.Error("GMT compliance failed for payment", "payment_id", id, "error", err)
			innerErr := workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to set payment status to failed. paymentID=", id, "err", innerErr)
				return innerErr
			}

			innerErr = workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, id, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to set transaction status to failed. paymentID=", id, "err", innerErr)
				return innerErr
			}

			return nil
		}
	}

	// launch payout and payin workflows in parallel
	selector := workflow.NewSelector(ctx)

	childPayInCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            fmt.Sprintf(payinWorkflowFmt, id),
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

	// don't set parent close policy
	childNotifyCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:            fmt.Sprintf(gmtNotifyCompleteFmt, id),
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON, // allow child workflow to continue running
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	})
	var we workflow.Execution
	err = workflow.ExecuteChildWorkflow(childNotifyCtx, gmt_workflows.GMTNotifyCompleted, id).GetChildWorkflowExecution().Get(childNotifyCtx, &we)
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
	accountsCtx = workflow.WithValue(accountsCtx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("paymentID=%s", paymentID),
	})

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var shouldPull bool
	err := workflow.ExecuteActivity(accountsCtx, a.ShouldPullFromAccount, paymentID).Get(ctx, &shouldPull)
	if err != nil {
		return err
	}

	if shouldPull {
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
			innerErr := workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateFailed).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to set transaction status to failed. paymentID=", paymentID, "err", innerErr)
				return innerErr
			}

			// Signal the Pay Out workflow that the payout has failed, it will know what to do.
			innerErr = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
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

				innerErr = workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateFailed).Get(ctx, nil)
				if innerErr != nil {
					logger.Error("Failed to set transaction status to failed. paymentID=", paymentID, "err", innerErr)
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
			return workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateFailed).Get(ctx, nil)
		}

		err = workflow.ExecuteActivity(accountsCtx, a.AddPayInTransfer, paymentID, accountTX.ID).Get(ctx, nil)
		if err != nil {
			return err
		}
	} else {
		// Just signal the payout workflow to go for it.
		err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:    paymentID,
			PayInSuccess: true,
		}).Get(ctx, nil)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(accountsCtx, a.AddPayInTransfer, paymentID, paymentID).Get(ctx, nil)
		if err != nil {
			return err
		}
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
		return workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateCompleted).Get(ctx, nil)
	}

	err = workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateFailed).Get(ctx, nil)
	if err != nil {
		return err
	}
	// Payout was unsuccessful. Rollback the account pull
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
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
	signalChanName       = "payment_signals"
	identityChanName     = "payment_identity_account_signals"
	payinWorkflowFmt     = "payment_pay_in_%s"
	payoutWorkflowFmt    = "payment_pay_out_%s"
	gmtNotifyCompleteFmt = "payment_gmt_notify_complete_%s"
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
	accountsCtx = workflow.WithValue(accountsCtx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("paymentID=%s", paymentID),
	})

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

	// Funding was a success now wait for the receiver to have all the relevant details
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
		WorkflowID:        fmt.Sprintf("payment_await_receiver_%s", paymentID),
	})

	var receiverReady bool
	err := workflow.ExecuteChildWorkflow(childCtx, AwaitReceiverWorkflow, paymentID).Get(childCtx, &receiverReady)
	if err != nil || !receiverReady {
		logger.Info("failed to wait for receiver to be ready to receive", "payment_id", paymentID, "err", err, "receiver_ready", receiverReady)
		// Signal the Pay In workflow that the payout has failed, it will know what to do.
		return workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:     paymentID,
			PayOutSuccess: false,
		}).Get(ctx, nil)
	}

	// Funding was a success, the receiver has all the relevant information, now payout
	var externalRef string
	err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
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

	err = workflow.ExecuteActivity(ctx, a.SetReceiveTransactionID, paymentID, txID).Get(ctx, nil)
	if err != nil {
		return err
	}

	// TODO: should we do 2-phase for the payout transaction?
	err = workflow.ExecuteActivity(ctx, a.UpdatePayoutTransactionState, paymentID, transactions.StateCompleted).Get(ctx, nil)
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
	ctx = workflow.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
		Context: fmt.Sprintf("paymentID=%s", paymentID),
	})

	err := workflow.ExecuteActivity(ctx, a.SetTransactionRefundState, paymentID, transactions.RefundStatePending).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction refundState", "Error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.RollbackPullFromAccount, paymentID).Get(ctx, nil)
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

	err = workflow.ExecuteActivity(ctx, a.SetTransactionRefundState, paymentID, transactions.RefundStateCompleted).Get(ctx, nil)
	if err != nil {
		logger.Error("error updating transaction refundState", "Error", err)
		return err
	}

	return nil
}

// AwaitReceiverWorkflow gets called from the PayoutWorkflow to await a user signing up and/or linking
// the identity associated with the payment, as well as a account that can receive.
func AwaitReceiverWorkflow(ctx workflow.Context, paymentID string) (bool, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var receiverReady bool
	err := workflow.ExecuteActivity(ctx, a.CheckReceiverReady, paymentID).Get(ctx, &receiverReady)
	if err != nil {
		return false, err
	}

	// The receiver can receive, nothing to do.
	if receiverReady {
		return true, nil
	}

	// Add the Workflow ref to the DB, so it can be signalled
	err = workflow.ExecuteActivity(ctx, a.AddIdentityWorkflowRef,
		paymentID, workflow.GetInfo(ctx).WorkflowExecution.ID, workflow.GetInfo(ctx).WorkflowExecution.RunID).Get(ctx, nil)
	if err != nil {
		return false, err
	}

	// Await the receiver being ready or a 24 hour timout
	var timeout bool
	selector := workflow.NewSelector(ctx)

	// Default to 24 hours for the user to link their accounts
	selector.AddFuture(workflow.NewTimer(ctx, time.Hour*24), func(f workflow.Future) {
		timeout = true
	})
	selector.AddReceive(workflow.GetSignalChannel(ctx, identityChanName), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
	})
	for {
		selector.Select(ctx)

		// Recheck that the receiver is ready whether it be timeout or signal
		err = workflow.ExecuteActivity(ctx, a.CheckReceiverReady, paymentID).Get(ctx, &receiverReady)
		if err != nil {
			logger.Error("failed to check receiver readiness", "err", err, "payment_id", paymentID)
		}

		if timeout && !receiverReady {
			err = workflow.ExecuteActivity(ctx, a.MarkWorkflowRefComplete,
				paymentID, workflow.GetInfo(ctx).WorkflowExecution.ID, workflow.GetInfo(ctx).WorkflowExecution.RunID).Get(ctx, nil)
			if err != nil {
				return false, err
			}

			logger.Info("Wait time expired for user to link identity or account for payment", "payment_id", paymentID)
			return false, nil
		}

		// Set the walletID if the identity is linked but not yet the linked account, so the workflow ID can be looked up.
		err = workflow.ExecuteActivity(ctx, a.SetWorkflowRefWalletID, paymentID).Get(ctx, nil)
		if err != nil {
			return false, err
		}

		// The receiver is ready, continue the workflow
		if receiverReady {
			break
		}
	}

	err = workflow.ExecuteActivity(ctx, a.MarkWorkflowRefComplete,
		paymentID, workflow.GetInfo(ctx).WorkflowExecution.ID, workflow.GetInfo(ctx).WorkflowExecution.RunID).Get(ctx, nil)
	if err != nil {
		return false, err
	}

	// Update the send transaction destination now that the user identity has been linked. We can look up the wallet address now.
	err = workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionDestination, paymentID).Get(ctx, nil)
	if err != nil {
		return false, err
	}

	// Receiver is ready, now run the compliance checks
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
	})

	// OFAC and compliance checks, now that the receiver is ready
	err = workflow.ExecuteChildWorkflow(childCtx, gmt_workflows.GMTComplianceChecksWorkflow, paymentID).Get(childCtx, nil)
	return err == nil, err
}
