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
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) error
	CreateSendUser(ctx context.Context, walletID string) error
	CreateReceiveBankAccount(ctx context.Context, args CreateReceiveBankAccountArgs) (*ReceiveBankAccount, error)
	GetReceiveBankAccount(ctx context.Context, id string) (*ReceiveBankAccount, error)
	CreateReceiveUser(ctx context.Context, args CreateReceiveUserArgs) (*ReceiveUser, error)
	GetReceiveUser(ctx context.Context, args GetReceiveUserArgs) (*ReceiveUser, error)
	CreateReceiveUserBankAccount(ctx context.Context, args CreateReceiveUserBankAccountArgs) (*ReceiveUserBankAccount, error)
	GetReceiveUserBankAccount(ctx context.Context, args GetReceiveUserBankAccountArgs) (*ReceiveUserBankAccount, error)
	GetBanks(ctx context.Context, countryCode string) ([]Bank, error)
}
