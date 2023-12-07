package pti

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateWallet(ctx context.Context, walletID string, currency currency.Currency) (Await, error)
}
