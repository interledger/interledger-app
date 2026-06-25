package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/xago"
	xago_external "gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/temporal"
)

// CrossProviderType identifies which cross-provider combination applies to a payment.
type CrossProviderType int

const (
	CrossProviderNone          CrossProviderType = iota
	CrossProviderGatehubToXago                   // EUR → ZAR
	CrossProviderXagoToGatehub                   // ZAR → EUR
)

// CheckCrossProviderType determines whether a payment is a cross-provider transfer.
// TODO check only what we need (CrossProviderGatehubToXago or CrossProviderXagoToGatehub)
func (a *Activity) CheckCrossProviderType(ctx context.Context, paymentID string) (CrossProviderType, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return CrossProviderNone, err
	}

	// TODO What are the payment types allowed for cross provider?
	// TODO What does p.SenderAccount == "" mean? How do we determine if it's an error?
	if p.Type != payments.TypePeer2Peer || p.SenderAccount == "" {
		// TODO check if this this the correct thing to return
		return CrossProviderNone, nil
	}

	senderAcc, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		// TODO check if this this the correct thing to return
		return CrossProviderNone, err
	}

	if p.ReceiverAccount != "" {
		receiverAcc, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
		if err != nil {
			// TODO check if this this the correct thing to return
			return CrossProviderNone, err
		}
		if senderAcc.Provider == gatehub.ProviderName && receiverAcc.Provider == xago.ProviderName {
			return CrossProviderGatehubToXago, nil
		}
		if senderAcc.Provider == xago.ProviderName && receiverAcc.Provider == gatehub.ProviderName {
			return CrossProviderXagoToGatehub, nil
		}
		return CrossProviderNone, nil
	}

	// TODO check if this logic is what we want
	// Fall back to currency check when receiver account not yet resolved.
	receiverCurrency := p.ReceiverAmount.Currency
	if senderAcc.Provider == gatehub.ProviderName && receiverCurrency == currency.ZAR {
		return CrossProviderGatehubToXago, nil
	}
	if senderAcc.Provider == xago.ProviderName && senderAcc.SendCurrency == currency.ZAR && receiverCurrency == currency.EUR {
		return CrossProviderXagoToGatehub, nil
	}

	return CrossProviderNone, nil
}

// CrossProviderGatehubReserve creates a pending Pacioli transfer:
// user.EUR → gatehub.EURClearingAccount (using p.SendTransactionID).
func (a *Activity) CrossProviderGatehubReserve(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}

	timeout := time.Hour * 24 * 365
	tx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              p.SendTransactionID,
			Amount:          p.SenderAmount.Value,
			DebitAccountID:  la.ID,
			CreditAccountID: gatehub.EURClearingAccount,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          gatehub.LedgerIDEUR,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExistsWithDifferentDebitAccountId ||
			tx[0].Code == pacioli.TransferExistsWithDifferentCreditAccountId {
			// Treat same-ID re-creation as idempotent success.
			return nil
		}
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits {
			return temporal.NewNonRetryableApplicationError("insufficient EUR balance", "insufficient_balance", nil)
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non-success Pacioli code (%s)", payments.ErrInternal, tx[0].Code.String())
		}
	}
	return nil
}

// CrossProviderXagoZARReserve creates a pending Pacioli transfer:
// user.ZAR → xago.ZARLiquidityAccount (using p.SendTransactionID).
func (a *Activity) CrossProviderXagoZARReserve(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}

	// TODO what value should we use for timeout?
	timeout := time.Hour * 24 * 365
	// TODO confirm whether we should create all transfers from the start (e.g. the clearing transfer)
	tx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              p.SendTransactionID,
			Amount:          p.SenderAmount.Value,
			DebitAccountID:  la.ID,
			CreditAccountID: xago.ZARLiquidityAccount,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          xago.LedgerIDZAR,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	// TODO maybe extract into a function
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExistsWithDifferentDebitAccountId ||
			tx[0].Code == pacioli.TransferExistsWithDifferentCreditAccountId {
			return nil
		}
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits {
			return temporal.NewNonRetryableApplicationError("insufficient ZAR balance", "insufficient_balance", nil)
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non-success Pacioli code (%s)", payments.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}

// CrossProviderGatehubTransferToOmnibus moves EUR from the sender's GateHub account to the GateHub omnibus.
func (a *Activity) CrossProviderGatehubTransferToOmnibus(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	externalTx, err := a.b.Gatehub().TransferUserToOmnibus(ctx, p.SenderAccount, p.SenderAmount)
	if err != nil {
		return "", fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return externalTx.ID, nil
}

// CrossProviderGatehubTransferFromOmnibus moves EUR from the GateHub omnibus to the given linked account.
// Used for Scenario 2 payout (omnibus → receiver) and Scenario 1 rollback (omnibus → sender).
func (a *Activity) CrossProviderGatehubTransferFromOmnibus(ctx context.Context, paymentID, receiverLinkedAccountID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	externalTx, err := a.b.Gatehub().TransferOmnibusToUser(ctx, receiverLinkedAccountID, p.ReceiverAmount)
	if err != nil {
		return "", fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return externalTx.ID, nil
}

// XagoConvertCurrencyActivity executes the Xago currency conversion and returns the convertID.
// For Scenario 1 (EUR→ZAR): pair=EURtoZAR, amount=senderAmount.
// For Scenario 2 (ZAR→EUR): pair=ZARtoEUR, amount=senderAmount.
func (a *Activity) XagoConvertCurrencyActivity(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	senderAcc, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return "", err
	}

	var pair xago_external.ConvertCurrencyPairEnum
	if senderAcc.Provider == gatehub.ProviderName {
		pair = xago_external.EURtoZAR
	} else {
		pair = xago_external.ZARtoEUR
	}

	// TODO maybe retry a few times
	resp, err := a.b.Xago().ConvertCurrency(ctx, pair, p.SenderAmount.Float64())
	if err != nil {
		return "", temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("Xago ConvertCurrency failed: %s", err),
			"XagoConvertError",
			err,
		)
	}

	return string(*resp), nil
}

// XagoConvertDetails holds the result of a completed Xago currency conversion.
type XagoConvertDetails struct {
	Complete        bool
	ConvertID       string
	Status          string
	Rate            string
	SendAmount      string
	SendFee         string
	ReceiveAmount   string
	ReceiveCurrency string
}

// XagoCheckConvertComplete polls GetConvertCurrencyDetails for a convertID.
// Returns complete=false when conversion is still in progress.
func (a *Activity) XagoCheckConvertComplete(ctx context.Context, convertID string) (XagoConvertDetails, error) {
	details, err := a.b.Xago().GetConvertCurrencyDetails(ctx, convertID)
	if err != nil {
		return XagoConvertDetails{}, err
	}

	status := strings.ToLower(details.Status)
	switch status {
	case "success":
		return XagoConvertDetails{
			Complete:        true,
			ConvertID:       details.ConvertID,
			Status:          details.Status,
			Rate:            details.Rate.String(),
			SendAmount:      details.SendAmount.String(),
			SendFee:         details.SendFee.String(),
			ReceiveAmount:   details.ReceiveAmount.String(),
			ReceiveCurrency: details.ReceiveCurrencyCode,
		}, nil
	case "pending":
		return XagoConvertDetails{Complete: false}, nil
	default:
		return XagoConvertDetails{}, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("Xago conversion failed with status: %s", details.Status),
			"XagoConversionFailed",
			nil,
		)
	}
}

// StoreActualFXRateAndAmount updates the payment's FX rate and receiver amount with the post-conversion actuals.
func (a *Activity) StoreActualFXRateAndAmount(ctx context.Context, paymentID string, details XagoConvertDetails) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	// TODO this conversion will lose precision
	receiveCurrency := currency.ParseCurrency(details.ReceiveCurrency)
	receiverAmt, err := currency.FromString(details.ReceiveAmount, receiveCurrency)
	if err != nil {
		return fmt.Errorf("parsing receiveAmount %q: %w", details.ReceiveAmount, err)
	}

	_, err = a.b.DB().ExecContext(ctx,
		"UPDATE payments SET fx_rate=$1, receiver_amount=$2, receiver_currency=$3 WHERE id=$4",
		details.Rate, receiverAmt.Value, receiverAmt.Currency.String(), p.ID,
	)
	return err
}

// PostGatehubToXagoTransfers atomically settles a Scenario 1 (EUR→ZAR) payment:
//  1. Posts the pending EUR reserve (p.SendTransactionID): user.EUR → gatehub.EURClearingAccount
//  2. Creates a posted transfer: xago.EURClearingAccount → xago.EUROpsAccount
//  3. Creates a posted transfer: xago.ZAROpsAccount → user.ZAR (using receiverTxID)
func (a *Activity) PostGatehubToXagoTransfers(ctx context.Context, paymentID, clearingTxID, receiverTxID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	receiverLA, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return err
	}

	// Post the EUR reserve pending transfer.
	postRes, err := a.b.Pacioli().PostTransfers(ctx, []string{p.SendTransactionID})
	if err != nil {
		return fmt.Errorf("%w posting EUR reserve: %s", payments.ErrInternal, err)
	}
	if len(postRes) > 0 && postRes[0].Code != 0 && postRes[0].Code != pacioli.TransferPendingTransferAlreadyPosted {
		return fmt.Errorf("%w non-success code posting reserve (%s)", payments.ErrInternal, postRes[0].Code.String())
	}

	// TODO check against the design doc
	// TODO Should we do this as pending true in the beginning and do PostTransfers here?
	createRes, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              clearingTxID,
			Amount:          p.SenderAmount.Value,
			DebitAccountID:  xago.EURClearingAccount,
			CreditAccountID: xago.EUROpsAccount,
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDEUR,
		},
		{
			ID:              receiverTxID,
			Amount:          p.ReceiverAmount.Value,
			DebitAccountID:  xago.ZAROpsAccount,
			CreditAccountID: receiverLA.ID,
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDZAR,
		},
	})
	if err != nil {
		return fmt.Errorf("%w creating GatehubToXago posted transfers: %s", payments.ErrInternal, err)
	}
	for _, r := range createRes {
		// TODO maybe create a function with a clearer description
		if r.Code != 0 &&
			r.Code != pacioli.TransferExistsWithDifferentDebitAccountId &&
			r.Code != pacioli.TransferExistsWithDifferentCreditAccountId {
			return fmt.Errorf("%w non-success code on GatehubToXago posted transfer (%s)", payments.ErrInternal, r.Code.String())
		}
	}
	return nil
}

// PostXagoToGatehubTransfers atomically settles a Scenario 2 (ZAR→EUR) payment:
//  1. Posts the pending ZAR reserve (p.SendTransactionID): user.ZAR → xago.ZARLiquidityAccount
//  2. Creates a posted transfer: xago.EUROpsAccount → xago.EURClearingAccount
//  3. Creates a posted transfer: gatehub.EURClearingAccount → user.EUR (using receiverTxID)
func (a *Activity) PostXagoToGatehubTransfers(ctx context.Context, paymentID, xagoOpsTxID, receiverTxID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	receiverLA, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return err
	}

	// Post the ZAR reserve pending transfer.
	postRes, err := a.b.Pacioli().PostTransfers(ctx, []string{p.SendTransactionID})
	if err != nil {
		return fmt.Errorf("%w posting ZAR reserve: %s", payments.ErrInternal, err)
	}
	// TODO extract and reuse logic for failed transactions
	if len(postRes) > 0 && postRes[0].Code != 0 && postRes[0].Code != pacioli.TransferPendingTransferAlreadyPosted {
		return fmt.Errorf("%w non-success code posting reserve (%s)", payments.ErrInternal, postRes[0].Code.String())
	}

	// TODO should the transactions be created at the beginning of the workflow?
	createRes, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              xagoOpsTxID,
			Amount:          p.ReceiverAmount.Value,
			DebitAccountID:  xago.EUROpsAccount,
			CreditAccountID: xago.EURClearingAccount,
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDEUR,
		},
		{
			ID:              receiverTxID,
			Amount:          p.ReceiverAmount.Value,
			DebitAccountID:  gatehub.EURClearingAccount,
			CreditAccountID: receiverLA.ID,
			Pending:         false,
			Code:            1,
			Ledger:          gatehub.LedgerIDEUR,
		},
	})
	if err != nil {
		return fmt.Errorf("%w creating XagoToGatehub posted transfers: %s", payments.ErrInternal, err)
	}
	for _, r := range createRes {
		// TODO extract and reuse logic for failed transactions
		if r.Code != 0 &&
			r.Code != pacioli.TransferExistsWithDifferentDebitAccountId &&
			r.Code != pacioli.TransferExistsWithDifferentCreditAccountId {
			return fmt.Errorf("%w non-success code on XagoToGatehub posted transfer (%s)", payments.ErrInternal, r.Code.String())
		}
	}
	return nil
}

// TODO review this CrossProviderGatehubRollbackTransfer

// CrossProviderGatehubRollbackTransfer moves EUR from the GateHub omnibus back to the sender.
// Called during rollback if the GateHub API transfer already happened (tracked via gatehub_transactions).
func (a *Activity) CrossProviderGatehubRollbackTransfer(ctx context.Context, paymentID string) error {
	// Check if GateHub API transfer was already made.
	var externalID sql.NullString
	// TODO check if possible to be here but the table is not filled yet
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_transactions WHERE payment_id = $1", paymentID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w querying gatehub_transactions: %s", payments.ErrInternal, err)
	}

	if !externalID.Valid || externalID.String == "" {
		// GateHub API transfer was not made, nothing to reverse.
		// TODO check if possible to be here but the table is not filled yet, and if so, maybe return error here to retry
		return nil
	}

	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	// Return EUR from omnibus back to sender.
	_, err = a.b.Gatehub().TransferOmnibusToUser(ctx, p.SenderAccount, p.SenderAmount)
	if err != nil {
		return fmt.Errorf("%w reversing GateHub transfer: %s", payments.ErrInternal, err)
	}

	return nil
}
