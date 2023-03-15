package gmt

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	StartUserOnboarding(ctx context.Context, walletID string) (Await, error)
}

type TransfersArgs struct {
	FromLinkedAccountID string `validate:"uuid"`
	ToLinkedAccountID   string `validate:"uuid"`
	Amount              currency.Amount
}

type Await func(context.Context, interface{}) error
