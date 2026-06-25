package ops

import (
	"time"

	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	"go.temporal.io/sdk/workflow"
)

// crossProviderGatehubToXagoPayIn handles Scenario 1 pay-in (EUR→ZAR, GateHub sender).
// Reserves EUR pending, calls GateHub API to move funds to omnibus, stores the external TX ID.
// The webhook wait happens in the corresponding pay-out.
func crossProviderGatehubToXagoPayIn(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {
	err := workflow.ExecuteActivity(ctx, a.CrossProviderGatehubReserve, paymentID).Get(ctx, nil)
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

// crossProviderXagoToGatehubPayIn handles Scenario 2 pay-in (ZAR→EUR, Xago sender).
// Reserves ZAR pending to the liquidity account.
func crossProviderXagoToGatehubPayIn(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {
	txID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	if err = workflow.ExecuteActivity(ctx, a.CrossProviderXagoZARReserve, paymentID).Get(ctx, nil); err != nil {
		return "", false, err
	}

	return txID, true, nil
}

// crossProviderGatehubToXagoPayOut handles Scenario 1 pay-out (EUR→ZAR, Xago ZAR receiver).
// Waits for the GateHub webhook confirming EUR is in omnibus, executes Xago EUR→ZAR conversion,
// polls for completion, and atomically posts all Pacioli transfers.
func crossProviderGatehubToXagoPayOut(ctx workflow.Context, a *Activity, paymentID string) (string, bool, error) {

	// Wait for GateHub webhook confirming sender's EUR reached omnibus.
	externalTransaction, ok, err := pollGatehubTransfer(ctx, a, paymentID)
	if err != nil || !ok {
		return "", false, err
	}

	// Generate stable UUIDs for the two new posted Pacioli transfers.
	clearingTxID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}
	receiverTxID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	// Execute Xago EUR→ZAR conversion (non-retryable; moves real money).
	var convertID string
	err = workflow.ExecuteActivity(ctx, a.XagoConvertCurrencyActivity, paymentID).Get(ctx, &convertID)
	if err != nil {
		return "", false, err
	}

	// Poll until conversion completes.
	// TODO extract and reuse
	for {
		var details XagoConvertDetails
		err = workflow.ExecuteActivity(ctx, a.XagoCheckConvertComplete, convertID).Get(ctx, &details)
		if err != nil {
			return "", false, err
		}
		if details.Complete {
			// TODO check what we need to store. What about the fees?
			err = workflow.ExecuteActivity(ctx, a.StoreActualFXRateAndAmount, paymentID, details).Get(ctx, nil)
			if err != nil {
				return "", false, err
			}
			break
		}
		// TODO review and refactor the polling for conversion complete
		_ = workflow.Sleep(ctx, 10*time.Second)
	}

	// Atomically post all Pacioli transfers.
	err = workflow.ExecuteActivity(ctx, a.PostGatehubToXagoTransfers, paymentID, clearingTxID, receiverTxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTransaction.ID, true, nil
}

// crossProviderXagoToGatehubPayOut handles Scenario 2 pay-out (ZAR→EUR, GateHub EUR receiver).
// Transfers EUR from GateHub omnibus to the receiver, waits for the webhook confirming delivery,
// executes Xago ZAR→EUR conversion, and atomically posts all Pacioli transfers.
func crossProviderXagoToGatehubPayOut(ctx workflow.Context, a *Activity, paymentID, receiverLinkedAccountID string) (string, bool, error) {
	logger := workflow.GetLogger(ctx)

	// Generate stable UUIDs for the two new posted Pacioli transfers.
	xagoOpsTxID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}
	receiverTxID, err := sideEffectUUID(ctx)
	if err != nil {
		return "", false, err
	}

	// Call GateHub API: omnibus → receiver.
	var externalTxID string
	err = workflow.ExecuteActivity(ctx, a.CrossProviderGatehubTransferFromOmnibus, paymentID, receiverLinkedAccountID).Get(ctx, &externalTxID)
	if err != nil {
		return "", false, err
	}

	// Store external TX ID so GetGatehubS2ReceiverTransfer can poll it.
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

	// TODO maybe extract and reuse this part of the code in GatehubToXagoPayOut
	// TODO is this the right order of operations or should we do conversion before gatehub part?
	// Execute Xago ZAR→EUR conversion (non-retryable; moves real money).
	var convertID string
	err = workflow.ExecuteActivity(ctx, a.XagoConvertCurrencyActivity, paymentID).Get(ctx, &convertID)
	if err != nil {
		return "", false, err
	}

	// Poll until conversion completes.
	// TODO extract and reuse
	for {
		var details XagoConvertDetails
		err = workflow.ExecuteActivity(ctx, a.XagoCheckConvertComplete, convertID).Get(ctx, &details)
		if err != nil {
			return "", false, err
		}
		if details.Complete {
			// TODO check what we need to store. What about the fees?
			err = workflow.ExecuteActivity(ctx, a.StoreActualFXRateAndAmount, paymentID, details).Get(ctx, nil)
			if err != nil {
				return "", false, err
			}
			break
		}

		// TODO review and refactor the polling for conversion complete
		_ = workflow.Sleep(ctx, 10*time.Second)
	}

	// Atomically post all Pacioli transfers.
	err = workflow.ExecuteActivity(ctx, a.PostXagoToGatehubTransfers, paymentID, xagoOpsTxID, receiverTxID).Get(ctx, nil)
	if err != nil {
		return "", false, err
	}

	return externalTxID, true, nil
}
