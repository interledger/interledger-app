package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
)

func CreatePaymentPointer(ctx context.Context, b Backends, w wallets.Wallet, assetCode string) error {
	if env.IsProd() {
		return nil
	}

	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", w.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if ppID != "" {
		return nil
	}

	ppID, err = b.External().CreatePaymentPointer(ctx, w, assetCode)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2)", w.ID, ppID)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	keys, err := b.Keys().List(ctx, w.ID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	for _, key := range keys {
		err := b.External().CreatePaymentPointerKey(ctx, ppID, key)
		if err != nil {
			return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
		}
	}

	return nil
}

func LookupWalletID(ctx context.Context, b Backends, paymentPointerID string) (string, error) {
	var wid string
	err := b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM rafiki_payment_pointers WHERE payment_pointer_id=$1", paymentPointerID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return wid, nil
}

func LookupPaymentPointerID(ctx context.Context, b Backends, walletID string) (string, error) {
	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", walletID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", rafiki.ErrNotFound, err)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return ppID, nil
}

func FundOutgoingPayment(ctx context.Context, b Backends, paymentID string) error {
	var eventID string
	err := b.DB().GetContext(ctx, &eventID, "SELECT event_id FROM rafiki_outgoing_payments WHERE payment_id=$1", paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	err = b.External().FundOutgoingPayment(ctx, eventID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func CreatePaymentPointerKey(ctx context.Context, b Backends, keyID string, walletID string) error {
	key, err := b.Keys().GetPublicKey(ctx, keyID, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	ppID, err := LookupPaymentPointerID(ctx, b, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	err = b.External().CreatePaymentPointerKey(ctx, ppID, *key)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func RevokePaymentPointerKey(ctx context.Context, b Backends, keyID string) error {
	err := b.External().RevokePaymentPointerKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("%w %s ", rafiki.ErrInternal, err)
	}
	return nil
}
