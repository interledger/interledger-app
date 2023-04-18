package ops

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

func CreateCard(ctx context.Context, b Backends, args verygoodsecurity.Card) (*verygoodsecurity.Card, error) {
	var card verygoodsecurity.Card
	sql := []string{
		"INSERT INTO very_good_security_card (card_token, expiry, card_security_code, wallet_id, last4, card_type) VALUES ($1, $2, $3, $4, $5, $6)",
		"ON CONFLICT (card_token, wallet_id) DO UPDATE SET (expiry, card_security_code, last4, card_type, updated_at) = (excluded.expiry, excluded.card_security_code, excluded.last4, excluded.card_type, now())",
		"RETURNING id, card_token, expiry, card_security_code, wallet_id, last4, card_type, created_at, updated_at;",
	}
	err := b.DB().GetContext(
		ctx,
		&card,
		strings.Join(sql, " "),
		args.Token,
		args.Expiry,
		args.CVV,
		args.WalletID,
		args.Last4,
		args.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", verygoodsecurity.ErrInternal, err)
	}

	return &card, nil
}
