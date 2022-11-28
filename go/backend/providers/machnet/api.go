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
	HandleEvent(ctx context.Context, event external.Event) error
	ValidateWebhook(ctx context.Context, payload []byte, base64Signature string) error
	CreateTransaction(ctx context.Context, args CreateTransactionArgs) (Await, error)
	StartSendUserKYC(ctx context.Context, walletID string) (Await, error)
	CreateReceiveBankAccount(ctx context.Context, args CreateReceiveBankAccountArgs) (*ReceiveBankAccount, error)
	GetReceiveBankAccount(ctx context.Context, id string) (*ReceiveBankAccount, error)
	CreateReceiveUser(ctx context.Context, args CreateReceiveUserArgs) (*ReceiveUser, error)
	GetReceiveUser(ctx context.Context, args GetReceiveUserArgs) (*ReceiveUser, error)
	CreateReceiveUserBankAccount(ctx context.Context, args CreateReceiveUserBankAccountArgs) (*ReceiveUserBankAccount, error)
	GetReceiveUserBankAccount(ctx context.Context, args GetReceiveUserBankAccountArgs) (*ReceiveUserBankAccount, error)
	GetBanks(ctx context.Context, countryCode string) ([]Bank, error)
	CreateWallet(ctx context.Context, args CreateWalletArgs) (*linkedaccounts.LinkedAccount, error)
	GetWallet(ctx context.Context, id string) (*Wallet, error)
	WithdrawFromWallet(ctx context.Context, args WithdrawFromWalletArgs) (*WalletWithdrawal, error)
	DeleteFundSource(ctx context.Context, linkedAccID string) (Await, error)
}
