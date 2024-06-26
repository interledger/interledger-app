package chimoney

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateWallet(ctx context.Context, walletID string) (Await, error)
	AddInterlocEmail(ctx context.Context, walletID, email string) (string, error)
	CreateDepositLink(ctx context.Context, walletID string, amt currency.Amount) (string, error)
}

type Await func(ctx context.Context, result interface{}) error
