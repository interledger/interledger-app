package astra

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

const (
	ProviderName = "astra"
	TypeCard     = "card"
)

type Await func(ctx context.Context, result interface{}) error

type CreateCardArgs struct {
	WalletID           string
	BasisTheoryTokenID string
}

type CardToAccountArgs struct {
	WalletID            string
	IdempotencyKey      string
	Name                string
	Amount              currency.Amount
	ClientCorrelationID string `validate:"len=8"` // Exactly 8 Chars
	DebitFeePercent     int
	CardID              string
}

type AccountToCardsArgs struct {
	WalletID        string
	IdempotencyKey  string
	Name            string
	Amount          currency.Amount
	DebitFeePercent int
	CardID          string
}
