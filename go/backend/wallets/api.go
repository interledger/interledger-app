package wallets

import (
	"context"
)

type Client interface {
	Create(ctx context.Context, args CreateArgs) (*Wallet, error)
	ForContext(ctx context.Context) (*Wallet, error)
	Get(ctx context.Context, id string) (*Wallet, error)
	List(ctx context.Context, userID string) ([]Wallet, error)
	SetWalletName(ctx context.Context, id, name string) (*Wallet, error)
}
