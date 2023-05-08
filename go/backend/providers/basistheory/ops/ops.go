package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
)

const (
	cardFields       = "id, wallet_id, token_id, expiration_month, expiration_year, tokenized_number, fingerprint, created_at, updated_at"
	insertCardFields = "wallet_id, token_id, expiration_month, expiration_year, tokenized_number, fingerprint"
)

func CreateCard(ctx context.Context, b Backends, tokenID, walletID string) (*basistheory.Card, error) {
	var card basistheory.Card
	err := b.DB().GetContext(
		ctx,
		&card,
		fmt.Sprintf("SELECT %s FROM basistheory_cards WHERE wallet_id=$1 AND token_id=$2;", cardFields),
		walletID,
		tokenID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}
	if card.ID != "" {
		return &card, nil
	}

	token, err := b.External().GetToken(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	cardData, err := external.ExtractCardDataFrom(token)
	if err != nil {
		return nil, fmt.Errorf("%w %s.", basistheory.ErrInternal, err)
	}

	err = b.DB().GetContext(
		ctx,
		&card,
		fmt.Sprintf("INSERT INTO basistheory_cards (%s) VALUES ($1, $2, $3, $4, $5, $6) RETURNING %s;", insertCardFields, cardFields),
		walletID,
		token.Id,
		cardData.ExpirationMonth,
		cardData.ExpirationYear,
		cardData.TokenizedNumber,
		token.GetFingerprint(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	return &card, nil
}

func GetCard(ctx context.Context, b Backends, id string) (*basistheory.Card, error) {
	var card basistheory.Card
	err := b.DB().GetContext(
		ctx,
		&card,
		fmt.Sprintf("SELECT %s FROM basistheory_cards WHERE id=$1;", cardFields),
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, basistheory.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	return &card, nil
}
