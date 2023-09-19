package rafiki

import (
	"context"
	"gitlab.com/fynbos/backend/wallets"
)

type Client interface {
	CreatePaymentPointer(ctx context.Context, address wallets.Wallet) error
}
