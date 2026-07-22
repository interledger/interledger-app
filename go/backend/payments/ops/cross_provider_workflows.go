package ops

import (
	"time"

	gatehub_external "github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
	"go.temporal.io/sdk/workflow"
)

// crossProviderGatehubToXagoPayIn handles pay-in for Gatehub (EUR) to Xago (ZAR).
// Reserves EUR pending, calls GateHub API to move funds to omnibus, stores the external TX ID.
// The webhook wait happens in the corresponding pay-out.
func crossProviderGatehubToXagoPayIn(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {
	err := workflow.ExecuteActivity(ctx, a.CrossProviderGatehubEURSenderReserve, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	var externalTxID string
	err = workflow.ExecuteActivity(ctx, a.CrossProviderGatehubTransferToOmnibus, paymentID).Get(ctx, &externalTxID)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubTransfer, paymentID, externalTxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTxID, true, nil
}

// crossProviderXagoToGatehubPayIn handles pay-in For Xago (ZAR) to Gatehub (EUR).
// Reserves ZAR pending to the liquidity account.
func crossProviderXagoToGatehubPayIn(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {
	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	if err = workflow.ExecuteActivity(ctx, a.CrossProviderXagoZARSenderReserve, paymentID).Get(ctx, nil); err != nil {
		return "", false, err
	}

	return txID, true, nil
}

// crossProviderGatehubToXagoPayOut handles pay-out for Gatehub (EUR) to Xago (ZAR).
// Waits for the Gatehub webhook confirming EUR is in omnibus, executes Xago EUR→ZAR conversion,
// polls for completion, and posts all Pacioli transfers.
func crossProviderGatehubToXagoPayOut(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {

	// Wait for GateHub webhook confirming sender's EUR reached omnibus.
	externalTransaction, ok, err := pollGatehubTransfer(ctx, a, paymentID)
	if err != nil || !ok {
		return "", false, err
	}

	// Execute Xago EUR→ZAR conversion (non-retryable; moves real money).
	var convertID string
	err = workflow.ExecuteActivity(ctx, a.XagoConvertCurrencyActivity, paymentID).Get(ctx, &convertID)
	if err != nil {
		return "", false, err
	}

	// Poll until conversion completes (the activity stores the actuals once complete).
	for {
		var done bool
		err = workflow.ExecuteActivity(ctx, a.XagoFinalizeConversion, paymentID, convertID).Get(ctx, &done)
		if err != nil {
			return "", false, err
		}
		if done {
			break
		}
		if err = workflow.Sleep(ctx, 10*time.Second); err != nil {
			return "", false, err
		}
	}

	// Atomically post all Pacioli transfers.
	err = workflow.ExecuteActivity(ctx, a.PostGatehubToXagoTransfers, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTransaction.ID, true, nil
}

// crossProviderXagoToGatehubPayOut handles pay-out for Xago (ZAR) to Gatehub (EUR).
// Transfers EUR from GateHub omnibus to the receiver, waits for the webhook confirming delivery,
// executes Xago ZAR→EUR conversion, and atomically posts all Pacioli transfers.
func crossProviderXagoToGatehubPayOut(ctx workflow.Context, a *Activity, paymentID, receiverLinkedAccountID string) (string, bool, error) {
	logger := workflow.GetLogger(ctx)

	// Call GateHub API: omnibus → receiver.
	var externalTxID string
	err := workflow.ExecuteActivity(ctx, a.CrossProviderGatehubTransferFromOmnibus, paymentID, receiverLinkedAccountID).Get(ctx, &externalTxID)
	if err != nil {
		return "", false, err
	}

	err = workflow.ExecuteActivity(ctx, a.SaveGatehubTransfer, paymentID, externalTxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	// TODO maybe extract in a function or reuse if possible
	// Wait for GateHub webhook confirming EUR reached the receiver.
	var externalTransaction gatehub_external.Transaction
	for {
		logger.Info("crossProviderXagoToGatehubPayOut: waiting for GateHub webhook")
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(workflow.NewTimer(ctx, 20*time.Minute), func(f workflow.Future) {
			logger.Info("GateHub webhook timer fired")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, gatehubNotifyChanName), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
		})
		selector.Select(ctx)

		err = workflow.ExecuteActivity(ctx, a.GetGatehubReceiverTransfer, paymentID).Get(ctx, &externalTransaction)
		if err != nil {
			return "", false, err
		}
		if externalTransaction.Status == gatehub_external.TransactionStatusCompleted {
			break
		} else if externalTransaction.Status == gatehub_external.TransactionStatusFailed {
			return "", false, nil
		}
	}

	// Atomically post all Pacioli transfers.
	err = workflow.ExecuteActivity(ctx, a.PostXagoToGatehubTransfers, paymentID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTxID, true, nil
}
