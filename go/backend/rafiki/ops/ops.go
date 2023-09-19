package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"
)

func CreatePaymentPointer(ctx context.Context, b Backends, w wallets.Wallet) error {
	var ppID string
	err := b.DB().GetContext(ctx, &ppID, "SELECT payment_pointer_id FROM rafiki_payment_pointers WHERE wallet_id=$1", w.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	if ppID != "" {
		return nil
	}

	ppID, err = b.External().CreatePaymentPointer(ctx, w)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2)", w.ID, ppID)
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}
