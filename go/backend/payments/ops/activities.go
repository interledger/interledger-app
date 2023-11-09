package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/providers/xago"
    "gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
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

	// No emails configured for withdrawal
	if payment.Type == payments.TypeWithdrawal {
		return nil
	}

	if payment.Type != payments.TypeWebMonetization {
		sendPaymentSentEmail(ctx, a.b, payment)
	}

	sendPaymentReceivedEmail(ctx, a.b, payment)

	return nil
}

func sendPaymentSentEmail(ctx context.Context, b Backends, payment *payments.Payment) {
	emails, greeting, err := b.Email().GetEmailsAndGreetingForWallet(ctx, payment.Sender.WalletID)
	if err != nil {
		log.Error("Failed to send payment sent email", zap.String("walletID", payment.Sender.WalletID), zap.String("paymentID", payment.ID))
		return
	}

	b.Email().PaymentSent(ctx, emails, greeting, payment)
}

func sendPaymentReceivedEmail(ctx context.Context, b Backends, payment *payments.Payment) {
	emails, greeting, err := b.Email().GetEmailsAndGreetingForWallet(ctx, payment.Receiver.WalletID)
	if err != nil {
		link, lerr := GetPaymentLinkByPaymentID(ctx, b, payment.ID)
		if lerr != nil {
			log.Error("Failed to send payment received email for payment link", zap.String("walletID", payment.Receiver.WalletID), zap.String("paymentID", payment.ID), zap.Error(err))
			return
		}

		id, ierr := b.KYC().GetIndividualDetails(ctx, link.ReceiverWalletID)
		if ierr != nil {
			log.Error("Failed to send payment received email for payment link. No kyc data.", zap.String("walletID", payment.Receiver.WalletID), zap.String("paymentID", payment.ID), zap.Error(err))
			return
		}

		emails = []sendgrid.Email{{
			Name:    fmt.Sprintf("%s %s", id.FirstName, id.LastName),
			Address: link.Email,
		}}
		greeting = strings.TrimSpace(fmt.Sprintf("Hello %s", id.FirstName)) + ","
	}

	b.Email().PaymentReceived(ctx, emails, greeting, payment)
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

	// No emails configured for withdrawal
	if payment.Type == payments.TypeWithdrawal {
		return nil
	}

	emails, greeting, err := a.b.Email().GetEmailsAndGreetingForWallet(ctx, payment.Sender.WalletID)
	if err != nil {
		log.Error("Failed to send payment failed email", zap.String("walletID", payment.Sender.WalletID))
		return nil
	}

	a.b.Email().PaymentFailed(ctx, emails, greeting)

	return nil
}

func (a *Activity) CheckPaymentSuccess(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
	}

	if p.Type == payments.TypeWithdrawal {
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

	return p.Type == payments.TypePeer2Peer || p.Type == payments.TypeRafikiPeer2Peer || p.Type == payments.TypeRafiki2External || p.Type == payments.TypeWithdrawal, nil
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

func (a *Activity) ReserveBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer {
		return nil
	}

	_, err = a.b.Xago().ReserveBalance(ctx, p.SenderAccount, p.SendTransactionID, p.SenderAmount)
	if errors.Is(err, xago.ErrInsufficientBalance) {
		return temporal.NewNonRetryableApplicationError("insufficient balance to service withdrawal", "insufficient_balance", err, "withdrawal", p.SenderAmount.Format())
	}

	return nil
}

func (a *Activity) AssignBalance(ctx context.Context, paymentID, txID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypePeer2Peer {
		return nil
	}

	_, err = a.b.Xago().AssignBalance(ctx, p.ReceiverAccount, txID, p.ReceiverAmount)
	return err
}

func (a *Activity) FinalizeBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer {
		return nil
	}

	return a.b.Xago().FinaliseReserve(ctx, p.SendTransactionID)
}

func (a *Activity) RollbackBalance(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if p.Type != payments.TypeWithdrawal && p.Type != payments.TypePeer2Peer {
		return nil
	}

	la, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if err != nil {
		return err
	}
	if la.Provider != xago.ProviderName {
		return nil
	}

	return a.b.Xago().RollbackReserve(ctx, p.SendTransactionID)
}
