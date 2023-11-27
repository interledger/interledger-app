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

func CreateWalletAddress(ctx context.Context, b Backends, w wallets.Wallet) error {
	if !env.IsDev() {
		return nil
	}

	var waID string
	err := b.DB().GetContext(ctx, &waID, "SELECT wallet_address_id FROM rafiki_wallet_addresses WHERE wallet_id=$1", w.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if waID != "" {
		return nil
	}

	waID, err = b.External().CreateWalletAddress(ctx, w)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_wallet_addresses (wallet_id, wallet_address_id) VALUES ($1, $2)", w.ID, waID)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}

func LookupWalletID(ctx context.Context, b Backends, walletAddressID string) (string, error) {
	var wid string
	err := b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM rafiki_wallet_addresses WHERE wallet_address_id=$1", walletAddressID)
	if err != nil {
		return "", fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return wid, nil
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
