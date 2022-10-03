package machnet

import "context"

type Client interface {
	GetUser(ctx context.Context, walletID string) (*User, error)
	CreateUser(ctx context.Context, args CreateArgs) (*User, error)
}
