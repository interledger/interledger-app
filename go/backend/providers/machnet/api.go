package machnet

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

type Client interface {
	External() external.Client
	GetKYCStatus(ctx context.Context, walletID string) (*UserKYC, error)
	GetUserByWalletID(ctx context.Context, walletID string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, args CreateArgs) (*User, error)
	GetWidgetToken(ctx context.Context, walletID string) (*WidgetToken, error)
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (Await, error)
	StartSendUserKYC(ctx context.Context, walletID string) (Await, error)
	CreateReceiveBankAccount(ctx context.Context, args CreateReceiveBankAccountArgs) (*ReceiveBankAccount, error)
	GetBanks(ctx context.Context, countryCode string) ([]Bank, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*linkedaccounts.LinkedAccount, error)
	GetWallet(ctx context.Context, id string) (*Wallet, error)
	WithdrawFromWallet(ctx context.Context, args WithdrawFromWalletArgs) (*WalletWithdrawal, error)
	DeleteFundSource(ctx context.Context, linkedAccID string) (Await, error)
}
