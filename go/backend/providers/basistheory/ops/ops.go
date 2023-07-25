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
	cardFields       = "id, wallet_id, token_id, expiration_month, expiration_year, tokenized_number, fingerprint, bin, pull_network, pull_enabled, pull_type, pull_country, push_network, push_enabled, push_type, push_availability, push_country, created_at, updated_at"
	insertCardFields = "wallet_id, token_id, expiration_month, expiration_year, tokenized_number, fingerprint, bin, pull_network, pull_enabled, pull_type, pull_country, push_network, push_enabled, push_type, push_availability, push_country"
)

func CreateCard(ctx context.Context, b Backends, args basistheory.CreateCardArgs) (*basistheory.Card, error) {
	var card basistheory.Card
	err := b.DB().GetContext(
		ctx,
		&card,
		fmt.Sprintf("SELECT %s FROM basistheory_cards WHERE wallet_id=$1 AND token_id=$2;", cardFields),
		args.WalletID,
		args.TokenID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}
	if card.ID != "" {
		return &card, nil
	}

	token, err := b.External().GetToken(ctx, args.TokenID)
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
		fmt.Sprintf("INSERT INTO basistheory_cards (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING %s;", insertCardFields, cardFields),
		args.WalletID,
		token.Id,
		cardData.ExpirationMonth,
		cardData.ExpirationYear,
		cardData.TokenizedNumber,
		token.GetFingerprint(),
		args.Bin,
		args.PullNetwork,
		args.PullEnabled,
		args.PullType,
		args.PullCountry,
		args.PushNetwork,
		args.PushEnabled,
		args.PushType,
		args.PushAvailability,
		args.PushCountry,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	return &card, nil
}

func UpdateCard(ctx context.Context, b Backends, new basistheory.UpdateCardArgs) (*basistheory.Card, error) {
	old, err := GetCard(ctx, b, new.ID)
	if err != nil {
		return nil, err
	}

	merged := *old

	noop := true
	if new.Bin != "" && new.Bin != old.Bin {
		merged.Bin = new.Bin
		noop = false
	}

	if new.PullCountry != "" && new.PullCountry != old.PullCountry {
		merged.PullCountry = new.PullCountry
		noop = false
	}

	if new.ExpirationMonth != "" && new.ExpirationMonth != old.ExpirationMonth {
		merged.ExpirationMonth = new.ExpirationMonth
		noop = false
	}

	if new.ExpirationYear != "" && new.ExpirationYear != old.ExpirationYear {
		merged.ExpirationYear = new.ExpirationYear
		noop = false
	}

	if new.Fingerprint != "" && new.Fingerprint != old.Fingerprint {
		merged.Fingerprint = new.Fingerprint
		noop = false
	}

	if new.PullNetwork != "" && new.PullNetwork != old.PullNetwork {
		merged.PullNetwork = new.PullNetwork
		noop = false
	}

	if new.PullEnabled != old.PullEnabled {
		merged.PullEnabled = new.PullEnabled
		noop = false
	}

	if new.PullType != "" && new.PullType != old.PullType {
		merged.PullType = new.PullType
		noop = false
	}

	if new.PushNetwork != "" && new.PushNetwork != old.PushNetwork {
		merged.PushNetwork = new.PushNetwork
		noop = false
	}

	if new.PushEnabled != old.PushEnabled {
		merged.PushEnabled = new.PushEnabled
		noop = false
	}

	if new.PushType != "" && new.PushType != old.PushType {
		merged.PushType = new.PushType
		noop = false
	}

	if new.PushAvailability != "" && new.PushAvailability != old.PushAvailability {
		merged.PushAvailability = new.PushAvailability
		noop = false
	}

	if new.PushCountry != "" && new.PushCountry != old.PushCountry {
		merged.PushCountry = new.PushCountry
		noop = false
	}

	if noop {
		return &merged, nil
	}

	result, err := b.DB().ExecContext(
		ctx,
		"UPDATE basistheory_cards SET expiration_month=$1, expiration_year=$2, fingerprint=$3, bin=$4, pull_network=$5, pull_enabled=$6, pull_type=$7, pull_country=$8, push_network=$9, push_enabled=$10, push_type=$11, push_availability=$12, push_country=$13, updated_at=now() WHERE id=$14;",
		merged.ExpirationMonth, merged.ExpirationYear, merged.Fingerprint, merged.Bin, merged.PullNetwork, merged.PullEnabled, merged.PullType, merged.PullCountry, merged.PushNetwork, merged.PushEnabled, merged.PushType, merged.PushAvailability, merged.PushCountry, merged.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	if rows, _ := result.RowsAffected(); rows < 1 {
		return nil, fmt.Errorf("%w No rows were updated.", basistheory.ErrInternal)
	}

	return &merged, nil
}

func ListCards(ctx context.Context, b Backends, limit uint) ([]basistheory.Card, error) {
	if limit < 1 {
		limit = 1000
	}
	var ret []basistheory.Card
	err := b.DB().SelectContext(ctx, &ret, fmt.Sprintf("SELECT %s FROM basistheory_cards LIMIT %d;", cardFields, limit))
	if err != nil {
		return nil, fmt.Errorf("%w %s", basistheory.ErrInternal, err)
	}

	return ret, nil
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
