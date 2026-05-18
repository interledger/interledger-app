package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/rafiki"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_external "gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b}
}

func (a *Activity) SetPaymentStateComplete(ctx context.Context, id string) error {
	payment, err := Lookup(ctx, a.b, id)
	if err != nil {
		if errors.Is(err, payments.ErrNotFound) {
			return temporal.NewApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	err = SetState(ctx, a.b, id, payments.StateCompleted)
	if err != nil {
		if errors.Is(err, payments.ErrInvalidStateTransition) {
			return temporal.NewApplicationError(err.Error(), "ErrInvalidStateTransition", err)
		}
		return err
	}

	if payment.Type == payments.TypeDeposit {
		la, err := a.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
		if err != nil {

			return nil
		}
		sourceAccountName := la.Name
		// flag(bradu): check pti.TypeCard
		if la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
			sourceAccountName = "Card ending " + strings.Replace(la.Mask, "*", "", -1)
		}
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
			sourceAccountName = fmt.Sprintf("%s %s", la.ReceiveNetwork, la.Mask)
		}

		a.b.Email().SendDepositReceivedEmail(ctx, payment.Sender.WalletID, payment.SenderAmount, sourceAccountName, payment.UpdatedAt.Format("02 Jan 2006"))
		return nil
	}

	if payment.Type == payments.TypeWithdrawal {
		la, err := a.b.LinkedAccounts().Get(ctx, payment.ReceiverAccount)
		if err != nil {

			return nil
		}

		destAccountName := la.Name
		if la.Provider == pti.ProviderName && la.Type == pti.TypeCard {
			destAccountName = "Card ending " + strings.Replace(la.Mask, "*", "", -1)
		}
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBank {
			destAccountName = fmt.Sprintf("%s %s", la.ReceiveNetwork, la.Mask)
		}

		a.b.Email().SendWithdrawalEmail(ctx, payment.Sender.WalletID, payment.SenderAmount, destAccountName, payment.UpdatedAt.Format("02 Jan 2006"))
		return nil
	}

	// No emails for web monetization
	if payments.TypeWebMonetization == payment.Type {
		return nil
	}

	a.b.Email().SendPaymentSentEmailV2(ctx, payment.Sender.WalletID, payment)

	a.b.Email().SendPaymentReceivedEmailV2(ctx, payment.Receiver.WalletID, payment)

	return nil
}

func (a *Activity) SetPaymentStateProcessing(ctx context.Context, id string) error {
	err := SetState(ctx, a.b, id, payments.StateProcessing)
	if err != nil {
		if errors.Is(err, payments.ErrInvalidStateTransition) {
			return temporal.NewApplicationError(err.Error(), "ErrInvalidStateTransition", err)
		}
		return err
	}

	return nil
}

func (a *Activity) SetPaymentStateFailed(ctx context.Context, id string) error {
	payment, err := Lookup(ctx, a.b, id)
	if err != nil {
		if errors.Is(err, payments.ErrNotFound) {
			return temporal.NewApplicationError(err.Error(), "ErrNotFound", err)
		}
		return err
	}

	err = SetState(ctx, a.b, id, payments.StateFailed)
	if err != nil {
		if errors.Is(err, payments.ErrInvalidStateTransition) {
			return temporal.NewApplicationError(err.Error(), "ErrInvalidStateTransition", err)
		}
		return err
	}

	if payment.Type == payments.TypeWithdrawal {
		a.b.Email().SendWithdrawalFailedEmail(ctx, payment.Sender.WalletID)
		return nil
	}

	if payment.Type == payments.TypeDeposit {
		a.b.Email().SendDepositFailedEmail(ctx, payment.Sender.WalletID)
		return nil
	}

	if payment.Type == payments.TypeWebMonetization {
		return nil
	}

	a.b.Email().SendPaymentFailedEmail(ctx, payment.Sender.WalletID)

	return nil
}

func (a *Activity) CheckPaymentSuccess(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	if p.Type == payments.TypeWithdrawal || p.Type == payments.TypeDeposit {
		if p.SendTransactionID == "" {
			return false, nil
		}

		senderTx, err := a.b.Transactions().GetTransaction(ctx, p.Sender.WalletID, p.SendTransactionID)
		if err != nil {
			return false, err
		}

		return senderTx.State == transactions.StateCompleted, nil
	}

	// If both transactions aren't present something failed.
	if p.ReceiveTransactionID == "" || p.SendTransactionID == "" || p.Receiver.WalletID == "" {
		return false, nil
	}

	senderTx, err := a.b.Transactions().GetTransaction(ctx, p.Sender.WalletID, p.SendTransactionID)
	if err != nil {
		return false, err
	}

	if senderTx.State == transactions.StateFailed {
		return false, nil
	}

	receiveTx, err := a.b.Transactions().GetTransaction(ctx, p.Receiver.WalletID, p.ReceiveTransactionID)
	if err != nil {
		return false, err
	}

	if receiveTx.State == transactions.StateFailed {
		return false, nil
	}

	return receiveTx.State == transactions.StateCompleted && senderTx.State == transactions.StateCompleted, nil
}

func (a *Activity) CheckReceiverReady(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	if p.ReceiverAccount != "" {
		return true, nil
	}

	if p.Receiver.WalletID == "" {
		return false, nil
	}

	// Check that the user has at least one account that can receive funds
	acc, err := a.b.LinkedAccounts().GetDefaultReceive(ctx, p.Receiver.WalletID, p.ReceiverAmount.Currency)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if p.ReceiverAccount == "" {
		_, err = update(ctx, a.b, payments.UpdateArgs{ID: paymentID, ReceiverAccount: acc.ID}, nil)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (a *Activity) SetWorkflowRefWalletID(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Receiver.WalletID == "" {
		return nil
	}

	_, err = a.b.DB().ExecContext(ctx, "UPDATE payments_workflow_refs SET wallet_id=$1, updated_at=now() WHERE payment_id=$2",
		p.Receiver.WalletID, paymentID)

	return err
}

func (a *Activity) AddIdentityWorkflowRef(ctx context.Context, paymentID, workflowID, runID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	_, err = a.b.DB().ExecContext(ctx, "INSERT INTO payments_workflow_refs (payment_id, identifier, wallet_id, workflow_id, workflow_run_id) VALUES ($1, $2, $3, $4, $5)",
		p.ID, p.Receiver.Identifier, p.Receiver.WalletID, workflowID, runID)

	return err
}

func (a *Activity) MarkWorkflowRefComplete(ctx context.Context, paymentID, workflowID, runID string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE payments_workflow_refs SET completed=true, updated_at=now() WHERE payment_id=$1 AND workflow_id=$2 AND workflow_run_id=$3",
		paymentID, workflowID, runID)

	return err
}

func (a *Activity) ShouldPullFromAccount(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	return p.Type == payments.TypePeer2Peer ||
		p.Type == payments.TypeRafikiPeer2Peer ||
		p.Type == payments.TypeRafiki2External ||
		p.Type == payments.TypeWithdrawal ||
		p.Type == payments.TypeWebMonetization, nil
}

func (a *Activity) ConfirmPaymentsEnginePayment(ctx context.Context, id string) ([]payments.RequiredActionType, error) {
	_, requiredActions, err := Confirm(ctx, a.b, id)
	return requiredActions, err
}

func (a *Activity) SignalRafikiPayIn(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeRafiki2External && p.Type != payments.TypeRafikiPeer2Peer {
		return nil
	}

	return a.b.Rafiki().FundOutgoingPayment(ctx, paymentID)
}

func (a *Activity) ShouldPushToAccount(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	return p.Type != payments.TypeRafiki2External, nil
}

func (a *Activity) LookupPayInAccount(ctx context.Context, paymentID string) (*linkedaccounts.LinkedAccount, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	return a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
}

func (a *Activity) LookupPayOutAccount(ctx context.Context, paymentID string) (*linkedaccounts.LinkedAccount, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	return a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
}

func (a *Activity) CheckXagoWithdrawalComplete(ctx context.Context, paymentID, externalID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	if p.Type != payments.TypeWithdrawal {
		return true, nil
	}

	wd, err := a.b.Xago().LookupWithdrawal(ctx, externalID)
	if err != nil {
		return false, err
	}

	if strings.EqualFold(wd.Status, "Success") {
		return true, nil
	}
	if strings.EqualFold(wd.Status, "rejected") || strings.EqualFold(wd.Status, "errored") {
		return false, temporal.NewNonRetryableApplicationError("withdrawal from xago failed", "external", nil, "status", wd.Status)
	}

	return false, nil
}

func (a *Activity) ReserveBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer && p.Type != payments.TypeRafikiPeer2Peer {
		return nil
	}

	timeout := time.Hour * 24 * 365 // Pending transfers must have a timeout.

	linkedAccount, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("sending linked account not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if linkedAccount.Provider == pti.ProviderName {
		_, err = a.b.PTI().ReserveBalance(ctx, linkedAccount.ID, p.SendTransactionID, p.SenderAmount, timeout)
		if errors.Is(err, pti.ErrInsufficientBalance) {
			return temporal.NewNonRetryableApplicationError("insufficient balance to service withdrawal", "insufficient_balance", err, "withdrawal", p.SenderAmount.Format())
		}
	} else if linkedAccount.Provider == xago.ProviderName {
		_, err = a.b.Xago().ReserveBalance(ctx, p.SenderAccount, p.SendTransactionID, p.SenderAmount, timeout)
		if errors.Is(err, xago.ErrInsufficientBalance) {
			return temporal.NewNonRetryableApplicationError("insufficient balance to service withdrawal", "insufficient_balance", err, "withdrawal", p.SenderAmount.Format())
		}
	} else if linkedAccount.Provider == gatehub.ProviderName {
		_, err = a.b.Gatehub().ReserveBalance(ctx, p.SenderAccount, p.SendTransactionID, p.SenderAmount, timeout)
		if errors.Is(err, gatehub.ErrInsufficientBalance) {
			return temporal.NewNonRetryableApplicationError("insufficient balance to service withdrawal", "insufficient_balance", err, "withdrawal", p.SenderAmount.Format())
		}
	} else if linkedAccount.Provider == chimoney.ProviderName {
		_, err = a.b.Chimoney().ReserveBalance(ctx, p.SenderAccount, p.SendTransactionID, p.SenderAmount, timeout)
		if errors.Is(err, chimoney.ErrInsufficientBalance) {
			return temporal.NewNonRetryableApplicationError("insufficient balance to service withdrawal", "insufficient_balance", err, "withdrawal", p.SenderAmount.Format())
		}
	}

	return err
}

func (a *Activity) AssignBalance(ctx context.Context, paymentID, txID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypePeer2Peer && p.Type != payments.TypeRafikiPeer2Peer && p.Type != payments.TypeDeposit && p.Type != payments.TypeWebMonetization {
		return nil
	}

	linkedAccount, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("sending linked account not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if linkedAccount.Provider == xago.ProviderName {
		_, err = a.b.Xago().AssignBalance(ctx, linkedAccount.ID, txID, p.ReceiverAmount)
	} else if linkedAccount.Provider == pti.ProviderName {
		_, err = a.b.PTI().AssignBalance(ctx, linkedAccount.ID, txID, p.ReceiverAmount)
	} else if linkedAccount.Provider == gatehub.ProviderName {
		_, err = a.b.Gatehub().AssignBalance(ctx, linkedAccount.ID, txID, p.ReceiverAmount)
	} else if linkedAccount.Provider == chimoney.ProviderName {
		_, err = a.b.Chimoney().AssignBalance(ctx, linkedAccount.ID, txID, p.ReceiverAmount)
	}

	return err
}

func (a *Activity) FinalizeBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type == payments.TypeWebMonetization {
		err = a.b.Rafiki().FinalizeWebMonetization(ctx, paymentID)
		if errors.Is(err, rafiki.ErrTimedOut) {
			return temporal.NewNonRetryableApplicationError("Web monetization payment has timed out", "ErrInternal", err)
		}
		if errors.Is(err, rafiki.ErrNotFound) {
			return temporal.NewNonRetryableApplicationError("Web monetization reserved transfer not found", "ErrNotFound", err)
		}
		if err != nil {
			return err
		}
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer && p.Type != payments.TypeRafikiPeer2Peer {
		return nil
	}

	linkedAccount, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("sending linked account not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if linkedAccount.Provider == xago.ProviderName {
		err = a.b.Xago().FinaliseReserve(ctx, p.SendTransactionID)
	} else if linkedAccount.Provider == pti.ProviderName {
		err = a.b.PTI().FinaliseReserve(ctx, p.SendTransactionID)
	} else if linkedAccount.Provider == gatehub.ProviderName {
		err = a.b.Gatehub().FinaliseReserve(ctx, p.SendTransactionID)
	} else if linkedAccount.Provider == chimoney.ProviderName {
		err = a.b.Chimoney().FinaliseReserve(ctx, p.SendTransactionID)
	}

	return err
}

func (a *Activity) RollbackBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type == payments.TypeWebMonetization {
		return a.b.Rafiki().RollbackWebMonetization(ctx, paymentID)
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer && p.Type != payments.TypeRafikiPeer2Peer {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}
	if la.Provider == xago.ProviderName {
		err = a.b.Xago().RollbackReserve(ctx, p.SendTransactionID)
	} else if la.Provider == pti.ProviderName {
		err = a.b.PTI().RollbackReserve(ctx, p.SendTransactionID)
	} else if la.Provider == gatehub.ProviderName {
		err = a.b.Gatehub().RollbackReserve(ctx, p.SendTransactionID)
	} else if la.Provider == chimoney.ProviderName {
		err = a.b.Chimoney().RollbackReserve(ctx, p.SendTransactionID)
	}

	return err
}

func (a *Activity) CheckReceiverLimits(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	w, err := lookupWallet(ctx, a.b, p.Receiver)
	if err != nil {
		return err
	}

	exceeded, _, err := a.b.Limits().ExceedsKYCLimits(ctx, w.ID, currency.FromUInt64(0, p.ReceiverAmount.Currency))
	if err != nil {
		return err
	}

	_, err = a.b.Wallets().SetExceededLimits(ctx, w.ID, exceeded)
	if err != nil {
		return err
	}

	if exceeded {
		a.b.Email().SendLimitsExceededEmail(ctx, w.ID)
	}

	return nil
}

func (a *Activity) GetPaymentType(ctx context.Context, paymentID string) (payments.Type, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return payments.TypeUnknown, err
	}

	return p.Type, nil
}

func (a *Activity) PTIWithdrawal(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	txID, err := a.b.PTI().WithdrawalFromWallet(ctx, pti.TransactionArgs{
		PaymentID:       paymentID,
		WalletID:        p.Sender.WalletID,
		Amount:          p.SenderAmount,
		LinkedAccountID: p.ReceiverAccount, // credit card should receive amount
	})
	if errors.Is(err, pti_external.ErrUnprocessableEntity) {
		return "", temporal.NewNonRetryableApplicationError("PTI unable to process withdrawal", "ErrUnprocessableEntity", err)
	}
	if err != nil {
		return "", err
	}

	return txID, err
}

func (a *Activity) PTIWithdrawalComplete(ctx context.Context, paymentID string, success bool) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}
	if la.Provider != pti.ProviderName {
		return nil
	}

	trl, err := a.b.Transactions().ListTransfers(ctx, p.SendTransactionID)
	if err != nil {
		return err
	}

	// There's only one transfer for a withdrawal
	var trxID string
	for _, tr := range trl {
		trxID = tr.ForeignID
	}

	state := pti.TransactionFeedbackAccepted
	if !success {
		state = pti.TransactionFeedbackCancelled
	}

	return a.b.PTI().UpdateTransactionStatus(ctx, pti.TransactionStatusArgs{
		PaymentID:     paymentID,
		TransactionID: trxID,
		Status:        state,
		Amount:        p.SenderAmount,
	})
}

func (a *Activity) PTIDepositComplete(ctx context.Context, paymentID, txID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeDeposit {
		return nil
	}

	return a.b.PTI().UpdateTransactionStatus(ctx, pti.TransactionStatusArgs{
		PaymentID:     paymentID,
		TransactionID: txID,
		Status:        pti.TransactionFeedbackSettled,
		Amount:        p.SenderAmount,
	})
}

func (a *Activity) ChimoneyTransfer(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	sender, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}

	receiver, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if err != nil {
		return err
	}

	err = a.b.Chimoney().Transfer(ctx, chimoney.TransferArgs{
		SendingWalletID:   sender.WalletID,
		ReceivingWalletID: receiver.WalletID,
		Amount:            p.SenderAmount,
	})
	if errors.Is(err, chimoney.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("chimoney wallet not found", "ErrNotFound", err)
	}

	return err
}

func (a *Activity) GatehubTransfer(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	externalTx, err := a.b.Gatehub().CreateTransfer(ctx, gatehub.CreateTransferArgs{
		SendingLinkedAccountID:   p.SenderAccount,
		ReceivingLinkedAccountID: p.ReceiverAccount,
		Amount:                   p.SenderAmount,
	})
	if errors.Is(err, gatehub.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Gatehub account not found", "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	return externalTx.ID, nil
}

func (a *Activity) SaveGatehubTransfer(ctx context.Context, paymentID, externalTransactionID string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT into gatehub_transactions (payment_id, external_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", paymentID, externalTransactionID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) GetGatehubTransfer(ctx context.Context, paymentID string) (*gatehub_external.Transaction, error) {
	var externalID string
	err := a.b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_transactions WHERE payment_id = $1;", paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, temporal.NewNonRetryableApplicationError("Payment not mapped to Gatehub transaction.", "ErrInternal", fmt.Errorf("%w Payment not mapped to Gatehub transaction.", payments.ErrInternal))
	}
	if err != nil {
		return nil, err
	}

	// // TODO Remove this ASAP
	// if strings.HasPrefix(externalID, "placeholder_") {
	// 	now := time.Now().UTC().Format(time.RFC3339)
	// 	return &gatehub_external.Transaction{
	// 		ID:          externalID,
	// 		Status:      gatehub_external.TransactionStatusCompleted,
	// 		CreatedAt:   now,
	// 		CompletedAt: now,
	// 	}, nil
	// }

	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	externalTrx, err := a.b.Gatehub().GetTransaction(ctx, p.Sender.WalletID, externalID)
	if err != nil {
		return nil, err
	}

	return externalTrx, nil
}
