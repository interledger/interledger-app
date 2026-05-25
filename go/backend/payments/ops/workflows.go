package ops

import (
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	pti_ops "gitlab.com/fynbos/backend/providers/pti/ops"
	"gitlab.com/fynbos/backend/providers/xago"
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

	err := workflow.ExecuteActivity(ctx, a.SetPaymentStateProcessing, id).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to set payment state to processing. paymentID=", id, "err", err)
		return err
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

	var childErr error
	for count := 0; count < 2; count++ {
		selector.Select(ctx)
		if err != nil {
			innerErr := workflow.ExecuteActivity(ctx, a.SetPaymentStateFailed, id).Get(ctx, nil)
			if innerErr != nil {
				logger.Error("Failed to set payment status to failed. paymentID=", id, "err", innerErr)
				return innerErr
			}

			// Wait for both to complete and do not return just yet
			childErr = err
		}
	}
	if childErr != nil {
		logger.Error("child workflow failed", "err", childErr)
		return nil
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

	return nil
}

func PayinWorkflow(ctx workflow.Context, paymentID string) error {
	var a *Activity
	var ptiActivity *pti_ops.Activity

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
		var la linkedaccounts.LinkedAccount
		err = workflow.ExecuteActivity(accountsCtx, a.LookupPayInAccount, paymentID).Get(ctx, &la)
		if err != nil {
			return err
		}

		var cpType CrossProviderType
		err = workflow.ExecuteActivity(accountsCtx, a.CheckCrossProviderType, paymentID).Get(ctx, &cpType)
		if err != nil {
			return err
		}

		var success bool
		var txID string
		switch la.Provider {
		case xago.ProviderName:
			if cpType == CrossProviderXagoToGatehub {
				txID, success, err = crossProviderXagoToGatehubPayIn(ctx, a, paymentID)
			} else {
				txID, success, err = xagoPayIn(ctx, a, paymentID)
			}
		case pti.ProviderName:
			txID, success, err = ptiPayIn(ctx, a, ptiActivity, paymentID, la.WalletID)
		case gatehub.ProviderName:
			if cpType == CrossProviderGatehubToXago {
				txID, success, err = crossProviderGatehubToXagoPayIn(ctx, a, paymentID)
			} else {
				txID, success, err = gatehubPayIn(ctx, a, paymentID, la.WalletID)
			}
		case chimoney.ProviderName:
			txID, success, err = chimoneyPayIn(ctx, a, paymentID, la.WalletID)
		default:
			return temporal.NewNonRetryableApplicationError("unsupported linked account provider", "InvalidArgument", nil, "provider", la.Provider)
		}
		if temporal_utils.IsNonRetryableError(err) || temporal_utils.IsMaxRetryError(err) {
			// Signal the Pay In workflow that the payout has failed, it will know what to do.
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

		// Notify payout tx of success or failure and it can get on with it
		err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:    paymentID,
			PayInSuccess: success,
		}).Get(ctx, nil)
		if err != nil {
			return err
		}

		if !success {
			// Mark transaction as a failure and stop the workflow
			return workflow.ExecuteActivity(ctx, a.UpdatePayInTransactionState, paymentID, transactions.StateFailed).Get(ctx, nil)
		}

		err = workflow.ExecuteActivity(accountsCtx, a.AddPayInTransfer, paymentID, txID).Get(ctx, nil)
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

	// Signal Rafiki that we pulled
	err = workflow.ExecuteActivity(ctx, a.SignalRafikiPayIn, paymentID, paymentID).Get(ctx, nil)
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
		err = workflow.ExecuteActivity(ctx, a.AddWithdrawalTransfer, paymentID).Get(ctx, nil)
		if err != nil {
			return err
		}

		err = workflow.ExecuteActivity(ctx, a.AddDepositTransfer, paymentID).Get(ctx, nil)
		if err != nil {
			return err
		}

		// Notify PTI of successful payout so complete the TX
		err = workflow.ExecuteActivity(ctx, a.PTIWithdrawalComplete, paymentID, true).Get(ctx, nil)
		if err != nil {
			return err
		}

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

func chimoneyPayIn(ctx workflow.Context, a *Activity, paymentID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	// Webmonetization is assumed to be Interledger to Interledger
	if pt != payments.TypePeer2Peer && pt != payments.TypeRafikiPeer2Peer && pt != payments.TypeWebMonetization {
		return "", false, temporal.NewNonRetryableApplicationError("incorrect payment type for chimoney pay in flow", "InvalidArgument", chimoney.ErrInternal, "paymentID", paymentID, "type", pt)
	}

	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.ReserveBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.ChimoneyTransfer, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return txID, true, nil
}

func chimoneyPayOut(ctx workflow.Context, a *Activity, paymentID, trxID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	// Webmonetization is assumed to be Interledger to Interledger
	if pt != payments.TypePeer2Peer && pt != payments.TypeRafikiPeer2Peer && pt != payments.TypeWebMonetization {
		return "", false, temporal.NewNonRetryableApplicationError("incorrect payment type for chimoney pay in flow", "InvalidArgument", gatehub.ErrInternal, "paymentID", paymentID, "type", pt)
	}

	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	// Commit balances for transfers and withdrawals
	err = workflow.ExecuteActivity(ctx, a.FinalizeBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// Assign balances for p2p receiver
	err = workflow.ExecuteActivity(ctx, a.AssignBalance, paymentID, trxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return txID, true, nil
}

func gatehubPayIn(ctx workflow.Context, a *Activity, paymentID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	// Webmonetization is assumed to be Interledger to Interledger
	if pt != payments.TypePeer2Peer && pt != payments.TypeRafikiPeer2Peer && pt != payments.TypeWebMonetization {
		return "", false, temporal.NewNonRetryableApplicationError("incorrect payment type for gatehub pay in flow", "InvalidArgument", gatehub.ErrInternal, "paymentID", paymentID, "type", pt)
	}

	err = workflow.ExecuteActivity(ctx, a.ReserveBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	var externalTransactionID string
	err = workflow.ExecuteActivity(ctx, a.GatehubTransfer, paymentID).Get(ctx, &externalTransactionID)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubTransfer, paymentID, externalTransactionID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTransactionID, true, nil
}

func gatehubPayOut(ctx workflow.Context, a *Activity, paymentID, trxID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	// Webmonetization is assumed to be Interledger to Interledger
	if pt != payments.TypePeer2Peer && pt != payments.TypeRafikiPeer2Peer && pt != payments.TypeWebMonetization {
		return "", false, temporal.NewNonRetryableApplicationError("incorrect payment type for gatehub pay in flow", "InvalidArgument", gatehub.ErrInternal, "paymentID", paymentID, "type", pt)
	}

	// Wait for gatehub completion, webhook or poll
	externalTransaction, ok, err := pollGatehubTransfer(ctx, a, paymentID)
	if err != nil || !ok {
		return "", false, err
	}

	// Commit balances for transfers and withdrawals
	err = workflow.ExecuteActivity(ctx, a.FinalizeBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// Assign balances for deposits and p2p receiver
	err = workflow.ExecuteActivity(ctx, a.AssignBalance, paymentID, trxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTransaction.ID, true, nil
}

func pollGatehubTransfer(ctx workflow.Context, a *Activity, paymentID string) (gatehub_external.Transaction, bool, error) {
	logger := workflow.GetLogger(ctx)
	var externalTransaction gatehub_external.Transaction
	for {
		logger.Info("pollGatehubTransfer entering selector loop", workflow.Now(ctx))
		selector := workflow.NewSelector(ctx)

		selector.AddFuture(workflow.NewTimer(ctx, 20*time.Minute), func(f workflow.Future) {
			logger.Info("Timer fired - 20 minutes elapsed")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, gatehubNotifyChanName), func(c workflow.ReceiveChannel, more bool) {
			logger.Info("***SIGNAL RECEIVED*** on gatehub notify channel", workflow.Now(ctx))
			c.Receive(ctx, nil)
		})

		selector.Select(ctx)
		logger.Info("Selector completed - checking transaction status", workflow.Now(ctx))

		err := workflow.ExecuteActivity(ctx, a.GetGatehubSenderTransfer, paymentID).Get(ctx, &externalTransaction)
		if err != nil {
			logger.Error("GetGatehubTransfer activity failed", "error", err)
			return gatehub_external.Transaction{}, false, err
		}

		logger.Info("Transaction status check", "status", externalTransaction.Status, "id", externalTransaction.ID)

		switch externalTransaction.Status {
		case gatehub_external.TransactionStatusCompleted:
			logger.Info("Transaction completed - breaking polling loop")
			return externalTransaction, true, nil
		case gatehub_external.TransactionStatusFailed:
			return gatehub_external.Transaction{}, false, nil
		}
	}
}

func ptiPayIn(ctx workflow.Context, a *Activity, ptiA *pti_ops.Activity, paymentID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.ReserveBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	if pt == payments.TypeWithdrawal {
		var txID string
		err = workflow.ExecuteActivity(ctx, a.PTIWithdrawal, paymentID).Get(ctx, &txID)
		var applicationError *temporal.ApplicationError
		// PTI lets us know they need more user info through 422 errors
		if errors.As(err, &applicationError) && applicationError.Type() == "ErrUnprocessableEntity" {
			innerErr := workflow.ExecuteActivity(ctx, ptiA.StartUserAssessment, walletID, pti.ScenarioWithdrawal).Get(ctx, nil)
			if innerErr != nil {
				return "", false, innerErr
			}

			innerErr = workflow.ExecuteActivity(ctx, ptiA.CheckUserAssessmentAccepted, walletID).Get(ctx, nil)
			if innerErr != nil {
				return "", false, innerErr
			}

			// now retry withdrawal
			innerErr = workflow.ExecuteActivity(ctx, a.PTIWithdrawal, paymentID).Get(ctx, &txID)
			if innerErr != nil {
				return "", false, err
			}
		} else if err != nil {
			return "", false, err
		}

		return txID, true, nil

		// Webmonetization is assumed to be Interledger to Interledger
	} else if pt == payments.TypePeer2Peer || pt == payments.TypeWebMonetization || pt == payments.TypeRafikiPeer2Peer {
		txID, err := sideEffectUUID(ctx)
		if err != nil {
			return "", false, err
		}

		err = workflow.ExecuteActivity(ctx, ptiA.CreateWalletTransfer, paymentID, txID).Get(ctx, nil)
		var applicationError *temporal.ApplicationError
		// PTI lets us know they need more user info through 422 errors
		if errors.As(err, &applicationError) && applicationError.Type() == "ErrUnprocessableEntity" {
			innerErr := workflow.ExecuteActivity(ctx, ptiA.StartUserAssessment, walletID, pti.ScenarioTransfer).Get(ctx, nil)
			if innerErr != nil {
				return "", false, innerErr
			}

			innerErr = workflow.ExecuteActivity(ctx, ptiA.CheckUserAssessmentAccepted, walletID).Get(ctx, nil)
			if innerErr != nil {
				return "", false, innerErr
			}

			// now retry transfer
			innerErr = workflow.ExecuteActivity(ctx, ptiA.CreateWalletTransfer, paymentID, txID).Get(ctx, nil)
			if innerErr != nil {
				return "", false, err
			}
		} else if err != nil {
			return "", false, err
		}

		err = workflow.ExecuteActivity(ctx, ptiA.SavePTITransaction, paymentID, txID).Get(ctx, nil)
		if err != nil {
			return "", false, err
		}

		for {
			var ptiTrx external.TransactionStatus
			err = workflow.ExecuteActivity(ctx, ptiA.GetPTITransaction, txID).Get(ctx, &ptiTrx)
			if err != nil {
				return "", false, err
			}

			if ptiTrx.Status == "AUTHORIZED" || ptiTrx.Status == "SETTLED" || ptiTrx.Status == "ACCEPTED" {
				break
			} else if ptiTrx.Status == "ERROR" || ptiTrx.Status == "REFUSED" || ptiTrx.Status == "CANCELED" {
				// Notify the Payout worklfow of failure
				innerErr := workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payoutWorkflowFmt, paymentID), "", signalChanName, PaySignal{
					PaymentID:    paymentID,
					PayInSuccess: false,
				}).Get(ctx, nil)
				return "", false, innerErr
			} else {
				// Lets not spin on their API
				_ = workflow.Sleep(ctx, time.Minute*10)
				continue
			}
		}

		return txID, true, nil
	}

	return "", false, temporal.NewNonRetryableApplicationError("incorrect payment type for pti pay in flow", "InvalidArgument", pti.ErrInternal, "paymentID", paymentID, "type", pt)
}

func xagoPayIn(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {
	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	// Check balance is sufficient. Debit Balance
	err = workflow.ExecuteActivity(ctx, a.ReserveBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return txID, true, nil
}

type PaySignal struct {
	PaymentID     string
	PayInSuccess  bool
	PayOutSuccess bool
}

const (
	signalChanName        = "payment_signals"
	identityChanName      = "payment_identity_account_signals"
	payinWorkflowFmt      = "payment_pay_in_%s"
	payoutWorkflowFmt     = "payment_pay_out_%s"
	gatehubNotifyChanName = "payment_gatehub_signals"
)

func PayoutWorkflow(ctx workflow.Context, paymentID string) error {
	var a *Activity
	var ptiActivity *pti_ops.Activity

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

	externalSuccess, doPayout, err := maybeAwaitRafikiPayout(ctx, a, paymentID)
	if err != nil {
		return err
	}

	// Payout done by rafiki, just signal success or failure.
	if !doPayout {
		err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:     paymentID,
			PayOutSuccess: externalSuccess,
		}).Get(ctx, nil)
		return err
	}

	// Funding was a success now wait for the receiver to have all the relevant details
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
		WorkflowID:        fmt.Sprintf("payment_await_receiver_%s", paymentID),
	})

	var receiverReady bool
	err = workflow.ExecuteChildWorkflow(childCtx, AwaitReceiverWorkflow, paymentID).Get(childCtx, &receiverReady)
	if err != nil || !receiverReady {
		logger.Info("failed to wait for receiver to be ready to receive", "payment_id", paymentID, "err", err, "receiver_ready", receiverReady)
		// Signal the Pay In workflow that the payout has failed, it will know what to do.
		return workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:     paymentID,
			PayOutSuccess: false,
		}).Get(ctx, nil)
	}

	// Generate the transactionID used for transactions service as well as Xago
	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return err
	}

	// Funding was a success, the receiver has all the relevant information, now payout

	// Switch here on provider of the receiver linkedAcc
	var receiverLA linkedaccounts.LinkedAccount
	err = workflow.ExecuteActivity(accountsCtx, a.LookupPayOutAccount, paymentID).Get(ctx, &receiverLA)
	if err != nil {
		return err
	}

	var cpType CrossProviderType
	err = workflow.ExecuteActivity(accountsCtx, a.CheckCrossProviderType, paymentID).Get(ctx, &cpType)
	if err != nil {
		return err
	}

	var success bool
	var externalTXID string
	switch receiverLA.Provider {
	case xago.ProviderName:
		if cpType == CrossProviderGatehubToXago {
			externalTXID, success, err = crossProviderGatehubToXagoPayOut(ctx, a, paymentID)
		} else {
			externalTXID, success, err = xagoPayOut(ctx, accountsCtx, a, paymentID, txID)
		}
	case pti.ProviderName:
		externalTXID, success, err = ptiPayOut(ctx, accountsCtx, a, ptiActivity, paymentID, txID, receiverLA.WalletID)
	case gatehub.ProviderName:
		if cpType == CrossProviderXagoToGatehub {
			externalTXID, success, err = crossProviderXagoToGatehubPayOut(ctx, a, paymentID, receiverLA.ID)
		} else {
			externalTXID, success, err = gatehubPayOut(ctx, a, paymentID, txID, receiverLA.WalletID)
		}
	case chimoney.ProviderName:
		externalTXID, success, err = chimoneyPayOut(ctx, a, paymentID, txID, receiverLA.WalletID)
	default:
		return temporal.NewNonRetryableApplicationError("unsupported linked account provider", "InvalidArgument", nil, "provider", receiverLA.Provider)
	}

	if temporal_utils.IsNonRetryableError(err) || temporal_utils.IsMaxRetryError(err) {
		// Signal the Pay In workflow that the payout has failed, it will know what to do.
		innerErr := workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
			PaymentID:     paymentID,
			PayOutSuccess: false,
		}).Get(ctx, nil)
		if innerErr != nil {
			logger.Error("failed to notify pay in workflow", "err", err)
		}
	}
	if err != nil {
		return err
	}

	// Notify payin tx of success or failure and it can get on with it
	err = workflow.SignalExternalWorkflow(ctx, fmt.Sprintf(payinWorkflowFmt, paymentID), "", signalChanName, PaySignal{
		PaymentID:     paymentID,
		PayOutSuccess: success,
	}).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Payout was unsuccessful, nothing more to do.
	if !success {
		return nil
	}

	// Create the incoming transaction
	err = workflow.ExecuteActivity(ctx, a.CreatePayoutTransaction, paymentID, txID, externalTXID).Get(ctx, &txID)
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

	err = workflow.ExecuteActivity(ctx, a.CheckReceiverLimits, paymentID).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func ptiPayOut(ctx, accountsCtx workflow.Context, a *Activity, ptiA *pti_ops.Activity, paymentID, trxID, walletID string) (string, bool, error) {
	var pt payments.Type
	err := workflow.ExecuteActivity(ctx, a.GetPaymentType, paymentID).Get(ctx, &pt)
	if err != nil {
		return "", false, err
	}

	var externalTxID string
	if pt == payments.TypePeer2Peer || pt == payments.TypeWebMonetization {
		var ptiTrx external.TransactionStatus
		err = workflow.ExecuteActivity(accountsCtx, ptiA.GetPTITransactionByPaymentID, paymentID).Get(ctx, &ptiTrx)
		if err != nil {
			return "", false, err
		}

		if ptiTrx.Status == "PENDING" || ptiTrx.Status == "MANUAL_REVIEW" {
			return "", false, temporal.NewApplicationError("PTI transaction is pending", "ErrInternal")
		} else if ptiTrx.Status == "ERROR" || ptiTrx.Status == "REFUSED" || ptiTrx.Status == "CANCELED" || ptiTrx.Status == "REFUNDED" || ptiTrx.Status == "CHARGED_BACK" {
			return "", false, temporal.NewNonRetryableApplicationError("PTI transaction failed", "ErrInternal", pti.ErrInternal)
		}

		externalTxID = ptiTrx.RequestID
	}

	// Commit balances for transfers and withdrawals
	err = workflow.ExecuteActivity(ctx, a.FinalizeBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// Assign balances for deposits and p2p receiver
	err = workflow.ExecuteActivity(ctx, a.AssignBalance, paymentID, trxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// TODO: commit all pacioli ledger entries for webmonetization

	return externalTxID, true, nil
}

func xagoPayOut(ctx, accountsCtx workflow.Context, a *Activity, paymentID, txID string) (string, bool, error) {
	var externalTX string
	err := workflow.ExecuteActivity(accountsCtx, a.WithdrawFromXagoBalance, paymentID).Get(accountsCtx, &externalTX)
	if err != nil {
		return "", false, err
	}

	// Wait for the withdrawal to complete on Xago side, if it is a withdrawal
	for {
		var complete bool
		err = workflow.ExecuteActivity(ctx, a.CheckXagoWithdrawalComplete, paymentID, externalTX).Get(ctx, &complete)
		if err != nil {
			return "", false, err
		}

		if complete {
			break
		}

		_ = workflow.Sleep(ctx, time.Hour) // Check again in an hour
	}

	// Finalize the debit on the balance
	err = workflow.ExecuteActivity(ctx, a.FinalizeBalance, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// Assign the new balance if transfer and not withdrawal
	err = workflow.ExecuteActivity(ctx, a.AssignBalance, paymentID, txID).Get(ctx, nil)
	if err != nil {
		return txID, false, err
	}

	return externalTX, true, nil
}

func maybeAwaitRafikiPayout(ctx workflow.Context, a *Activity, paymentID string) (success bool, doPayout bool, err error) {
	err = workflow.ExecuteActivity(ctx, a.ShouldPushToAccount, paymentID).Get(ctx, &doPayout)
	if err != nil {
		return
	}

	if doPayout {
		return true, true, nil
	}

	// Wait for rafiki webhook to tell us the payment out was successful or failed
	var signal PaySignal
	signalChan := workflow.GetSignalChannel(ctx, signalChanName)
	for {
		signalChan.Receive(ctx, &signal)
		if signal.PaymentID != paymentID {
			continue
		}

		success = signal.PayOutSuccess
		return
	}
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

	// TODO check if this voids the Xago->Gatehub cross provider payIn transfer in pacioli
	err = workflow.ExecuteActivity(ctx, a.RollbackBalance, paymentID).Get(ctx, nil)
	if err != nil {
		logger.Error("error rolling back balance reserve", "Error", err)
		return err
	}

	// For cross-provider GateHub→Xago: if the GateHub API transfer already happened,
	// move EUR from omnibus back to the sender.
	var cpType CrossProviderType
	err = workflow.ExecuteActivity(ctx, a.CheckCrossProviderType, paymentID).Get(ctx, &cpType)
	if err != nil {
		logger.Error("error checking cross provider type for rollback", "Error", err)
		return err
	}
	if cpType == CrossProviderGatehubToXago {
		err = workflow.ExecuteActivity(ctx, a.CrossProviderGatehubRollbackTransfer, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("error rolling back GateHub cross-provider transfer", "Error", err)
			return err
		}
	}

	// Rollback PTI withdrawal
	err = workflow.ExecuteActivity(ctx, a.PTIWithdrawalComplete, paymentID, false).Get(ctx, nil)
	if err != nil {
		logger.Error("error PTI withdrawal rolling back balance reserve", "Error", err)
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

	return true, nil
}
