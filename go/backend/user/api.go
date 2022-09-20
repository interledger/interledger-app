package user

import "context"

type Client interface {
	UserForCookie(ctx context.Context, cookie string) (*User, error)
	UserForContext(ctx context.Context) (*User, error)
	WalletForContext(ctx context.Context) (*Wallet, error)
	CreateNewWallet(ctx context.Context, userID, walletName string) (*Wallet, error)
	ListWallets(ctx context.Context, userID string) ([]Wallet, error)
	GetWallet(ctx context.Context, userID, id string) (*Wallet, error)
}
