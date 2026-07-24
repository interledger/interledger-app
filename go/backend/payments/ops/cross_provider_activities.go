package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	xago_external "github.com/interledger/interledger-app/go/backend/providers/xago/external"
	"github.com/interledger/interledger-app/go/pacioli"
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

// CrossProviderGatehubEURSenderReserve creates a pending Pacioli transfer:
// gatehub.user.EURAccount → gatehub.EURClearingAccount (using p.SendTransactionID).
func (a *Activity) CrossProviderGatehubEURSenderReserve(ctx context.Context, paymentID string) error {
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

// CrossProviderXagoZARSenderReserve creates a pending Pacioli transfer:
// xago.user.ZARAccount → xago.ZARLiquidityAccount (using p.SendTransactionID).
func (a *Activity) CrossProviderXagoZARSenderReserve(ctx context.Context, paymentID string) error {
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
// Used for Xago to Gatehub (omnibus → receiver) and Gatehub to Xago rollback (omnibus → sender).
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

// XagoFinalizeConversion reports whether the Xago conversion has completed, persisting its results
//   - returns false while the conversion is still pending
//   - returns a non-retryable error if the conversion failed
func (a *Activity) XagoFinalizeConversion(ctx context.Context, paymentID, convertID string) (bool, error) {
	details, done, err := xagoCheckConvertComplete(ctx, a.b, convertID)
	if err != nil {
		return false, err
	}
	if !done {
		// Conversion still pending
		return false, nil
	}

	if err := storeActualFXRateAndAmount(ctx, a.b, paymentID, details); err != nil {
		return false, err
	}
	if err := storeXagoConversion(ctx, a.b, paymentID, details); err != nil {
		return false, err
	}
	return true, nil
}

// StoreXagoConversionEstimation snapshots the payment's FX draft into xago_currency_conversion_estimations
// as the estimate committed at execution time. It returns early for payments that are not a GateHub <-> Xago
// transfer (either direction), and is idempotent via the payment_id conflict on the estimations table.
func (a *Activity) StoreXagoConversionEstimation(ctx context.Context, paymentID string) error {
	payment, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if payment.SenderAccount == "" || payment.ReceiverAccount == "" {
		return nil
	}

	senderAccount, err := a.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return err
	}

	receiverAccount, err := a.b.LinkedAccounts().Get(ctx, payment.ReceiverAccount)
	if err != nil {
		return err
	}

	// Both directions of the GateHub <-> Xago pair carry an FX estimate to snapshot.
	if !isXagoGatehubPair(senderAccount.Provider, receiverAccount.Provider) {
		return nil
	}

	_, err = a.b.DB().ExecContext(ctx,
		`INSERT INTO xago_currency_conversion_estimations (
			payment_id, estimated_rate, send_amount, send_currency_code, receive_amount, receive_currency_code, raw_data
		)
		SELECT payment_id, estimated_rate, send_amount, send_currency_code, receive_amount, receive_currency_code, raw_data
		FROM xago_currency_conversion_estimation_drafts
		WHERE payment_id = $1`,
		paymentID,
	)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("storing xago conversion estimation for payment %s: %w", paymentID, err)
	}
	return nil
}

// PostGatehubToXagoTransfers posts Gatehub (EUR) to Xago (ZAR) transfers:
//  1. Posts the pending EUR reserve (p.SendTransactionID): gatehub.user.EUR → gatehub.EURClearingAccount
//  2. Creates a posted transfer: xago.EURClearingAccount → xago.EUROpsAccount
//  3. Creates a posted transfer: xago.ZAROpsAccount → xago.ZARLiquidityAccount
//  4. Creates a posted transfer: xago.ZARLiquidityAccount → xago.user1.ZARAccount
func (a *Activity) PostGatehubToXagoTransfers(ctx context.Context, paymentID string) error {
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

	createRes, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              uuid.NewString(),
			Amount:          p.SenderAmount.Value,
			DebitAccountID:  xago.EURClearingAccount,
			CreditAccountID: xago.EUROpsAccount,
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDEUR,
		},
		{
			ID:              uuid.NewString(),
			Amount:          p.ReceiverAmount.Value,
			DebitAccountID:  xago.ZAROpsAccount,
			CreditAccountID: xago.ZARLiquidityAccount,
			Pending:         false,
			Code:            1,
			Ledger:          xago.LedgerIDZAR,
		},
		{
			ID:              uuid.NewString(),
			Amount:          p.ReceiverAmount.Value,
			DebitAccountID:  xago.ZARLiquidityAccount,
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

// PostXagoToGatehubTransfers posts Xago (ZAR) to Gatehub (EUR) transfers in pacioli:
//  1. Posts the pending ZAR reserve (p.SendTransactionID): xago.user.ZARAccount → xago.ZARLiquidityAccount
//  2. Creates a posted transfer: gatehub.EURClearingAccount → gatehub.user.EURAccount (using receiverTxID)
func (a *Activity) PostXagoToGatehubTransfers(ctx context.Context, paymentID string) error {
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
			ID:              uuid.NewString(),
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

func (a *Activity) AddXagoTravelRuleRecord(ctx context.Context, paymentID string) error {
	payment, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if payment.SenderAccount == "" || payment.ReceiverAccount == "" {
		return nil
	}

	senderAccount, err := a.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return err
	}
	if senderAccount.Provider != gatehub.ProviderName {
		return nil
	}

	receiverAccount, err := a.b.LinkedAccounts().Get(ctx, payment.ReceiverAccount)
	if err != nil {
		return err
	}
	if receiverAccount.Provider != xago.ProviderName {
		return nil
	}

	err = a.b.Xago().InsertTravelRuleRecord(ctx, xago.TravelRuleRecordArgs{
		PaymentID:        payment.ID,
		SenderWalletID:   payment.Sender.WalletID,
		ReceiverWalletID: payment.Receiver.WalletID,
	})
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	return err
}

// SetPayoutTransactionXagoFX records the applied FX rate, surcharge and target amount on a payout
// transaction. It is a no-op when the payment has no Xago FX data.
func (a *Activity) SetPayoutTransactionXagoFX(ctx context.Context, paymentID, txID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	return setTransactionXagoFX(ctx, a.b, paymentID, txID, p.SenderAmount)
}

// SetPayInTransactionXagoFX records the applied FX rate, surcharge and target amount on a payin
// transaction. It is a no-op when the payment has no Xago FX data.
func (a *Activity) SetPayInTransactionXagoFX(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}
	if p.SendTransactionID == "" {
		return nil
	}

	return setTransactionXagoFX(ctx, a.b, paymentID, p.SendTransactionID, p.ReceiverAmount)
}
