package user

import (
	"context"

	"gitlab.com/fynbos/backend/db"
)

type Client interface {
	UserForCookie(ctx context.Context, cookie string) (*User, error)
	UserForContext(ctx context.Context) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context, walletID string) ([]User, error)
	WalletForContext(ctx context.Context) (*Wallet, error)
	CreateNewWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error)
	ListWallets(ctx context.Context, userID string) ([]Wallet, error)
	GetWallet(ctx context.Context, id string) (*Wallet, error)
	ListAllWallets(ctx context.Context, pagination db.Pagination) ([]Wallet, error)
	SetWalletName(ctx context.Context, id, name string) error
}
