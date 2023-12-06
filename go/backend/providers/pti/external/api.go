package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, args CreateUserArgs) (string, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error)
	GetWallet(ctx context.Context, id string) (*Wallet, error)
}
