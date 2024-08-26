package chimoney

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
)

type Client interface {
	CreateWallet(ctx context.Context, walletID string) (Await, error)
	AddInterlocEmail(ctx context.Context, walletID, email string) (*linkedaccounts.LinkedAccount, error)
	GetInterlocEmail(ctx context.Context, walletID string) (string, error)
	CreateDepositLink(ctx context.Context, walletID string, amt currency.Amount) (string, error)
	CreateDeposit(ctx context.Context, issueID string) (Await, error)
	ExecuteWithdraw(ctx context.Context, walletID string, transactionID string) error
	Transfer(ctx context.Context, args TransferArgs) error
	GetKYCWidget(ctx context.Context, walletID string) (string, error)
	WatchForSuccessfulKYC(ctx context.Context, walletID string) error

	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*Balance, error)
	RollbackReserve(ctx context.Context, txID string) error
}

type Await func(ctx context.Context, result interface{}) error
