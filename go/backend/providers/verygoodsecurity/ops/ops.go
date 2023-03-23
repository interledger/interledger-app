package ops

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

func CreateCard(ctx context.Context, b Backends, args verygoodsecurity.Card) (*verygoodsecurity.Card, error) {
	var card verygoodsecurity.Card
	err := b.DB().GetContext(
		ctx,
		&card,
		"INSERT INTO very_good_security_card (card_token, expiry, card_security_code, wallet_id, last4, card_type) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, card_token, expiry, card_security_code, wallet_id, last4, card_type, created_at, updated_at;",
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
