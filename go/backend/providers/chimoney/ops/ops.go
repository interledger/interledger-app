package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/providers/chimoney"
)

func UpsertInterlocEmail(ctx context.Context, b Backends, walletID, email string) error {
	_, err := b.DB().ExecContext(ctx, `INSERT INTO chi_money_interac_emails (wallet_id, email)
		VALUES ($1, $2)
		ON CONFLICT (wallet_id) 
		DO UPDATE SET 
			email = EXCLUDED.email,
			updated_at = NOW()`, walletID, email)
	if err != nil {
		return fmt.Errorf("%w failed to insert interloc email: %s", chimoney.ErrInternal, err)
	}

	return nil
}
