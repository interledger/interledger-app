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
	CreateReceiveBankAccount(ctx context.Context, args CreateReceiveBankAccountArgs) (*ReceiveBankAccount, error)
	GetReceiveBankAccount(ctx context.Context, id string) (*ReceiveBankAccount, error)
	CreateReceiveUser(ctx context.Context, args CreateReceiveUserArgs) (*ReceiveUser, error)
	GetReceiveUserByReceiveWalletID(ctx context.Context, receiveWalletID string) (*ReceiveUser, error)
	CreateReceiveUserBankAccount(ctx context.Context, args CreateReceiveUserBankAccountArgs) (*ReceiveUserBankAccount, error)
	GetReceiveUserAccountByReceiveAccountID(ctx context.Context, receiveAccountID string) (*ReceiveUserBankAccount, error)
}
