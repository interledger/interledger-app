package machnet

import (
	"context"

	"gitlab.com/fynbos/backend/providers/machnet/external"
)

type Client interface {
	GetUserByWalletID(ctx context.Context, walletID string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, args CreateArgs) (*User, error)
	GetWidgetToken(ctx context.Context, walletID string) (*WidgetToken, error)
	HandleEvent(ctx context.Context, event external.Event) error
	CreateReceiveAccount(ctx context.Context, args CreateReceiveAccountArgs) (*ReceiveAccount, error)
	GetReceiveAccount(ctx context.Context, id string) (*ReceiveAccount, error)
	CreateReceiveUser(ctx context.Context, args CreateReceiveUserArgs) (*ReceiveUser, error)
	GetReceiveUserByReceiveWalletID(ctx context.Context, receiveWalletID string) (*ReceiveUser, error)
	CreateReceiveUserAccount(ctx context.Context, args CreateReceiveUserAccountArgs) (*ReceiveUserAccount, error)
	GetReceiveUserAccountByReceiveAccountID(ctx context.Context, receiveAccountID string) (*ReceiveUserAccount, error)
}
