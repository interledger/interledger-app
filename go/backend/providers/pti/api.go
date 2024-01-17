package pti

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateWallet(ctx context.Context, walletID string, currency currency.Currency) (Await, error)
	GetWallet(ctx context.Context, linkedAccountID string) (*Wallet, error)
	DepositToWallet(ctx context.Context, args TransactionArgs) (string, error)
	WithdrawalFromWallet(ctx context.Context, args TransactionArgs) (string, error)
	UpdateTransactionStatus(ctx context.Context, args TransactionStatusArgs) error
}
