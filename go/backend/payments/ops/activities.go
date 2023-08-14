package ops

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/payments"
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

	senderWallet, err := a.b.Identities().GetByIdentifier(ctx, payment.Sender.Identifier)
	if err != nil {
		return err
	}
	a.b.Email().SendPaymentSentEmailV2(ctx, senderWallet.WalletID, payment)

	receiverWallet, err := a.b.Identities().GetByIdentifier(ctx, payment.Receiver.Identifier)
	if err != nil {
		return err
	}
	a.b.Email().SendPaymentReceivedEmailV2(ctx, receiverWallet.WalletID, payment)

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

	senderWallet, err := a.b.Identities().GetByIdentifier(ctx, payment.Sender.Identifier)
	if err != nil {
		return err
	}
	a.b.Email().SendPaymentFailedEmail(ctx, senderWallet.WalletID)

	return nil
}
