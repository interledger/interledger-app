package ops

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
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

	a.b.Email().SendPaymentFailedEmail(ctx, payment.Sender.WalletID)

	return nil
}

func (a *Activity) CheckPaymentSuccess(ctx context.Context, paymentID string) (bool, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return false, err
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

	w, err := lookupWallet(ctx, a.b, p.Receiver)
	if errors.Is(err, identities.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check that the user has at least one account that can receive funds
	acc, err := a.b.LinkedAccounts().GetDefaultReceive(ctx, w.ID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if p.ReceiverAccount == "" {
		_, err = update(ctx, a.b, payments.UpdateArgs{ID: paymentID, ReceiverAccount: acc.ID})
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

	return p.Type == payments.TypePeer2Peer, nil
}
