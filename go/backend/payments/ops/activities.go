package ops

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
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

	senderWallet, err := lookupWallet(ctx, a.b, payment.Sender)
	if err != nil {
		return err
	}
	a.b.Email().SendPaymentFailedEmail(ctx, senderWallet.ID)

	return nil
}

func lookupWallet(ctx context.Context, b Backends, identity payments.Identity) (*wallets.Wallet, error) {
	var resp *wallets.Wallet
	var err error
	switch identity.Type {
	case payments.IdentityTypeWalletID:
		resp, err = b.Wallets().Get(ctx, identity.Identifier)
	case payments.IdentityTypeWalletURL:
		resp, err = b.Wallets().GetFromAddress(ctx, identity.Identifier)
	case payments.IdentityTypeTwitter:
		var id *identities.Identity
		id, err = b.Identities().GetByIdentifier(ctx, identity.Identifier)
		if err != nil {
			return nil, err
		}
		if id.Platform != identities.PlatformTwitter {
			return nil, fmt.Errorf("identifier (%s) type mismatch expected (%s) got (%s)", identity.Identifier, identities.PlatformTwitter)
		}
		resp, err = b.Wallets().Get(ctx, id.WalletID)
	default:
		return nil, fmt.Errorf("unknown identity type %s", identity.Type)
	}
	return resp, err
}
