package gatehub

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateUser(ctx context.Context, walletID string) (Await, error)
	GetOnboardingWidget(ctx context.Context, walletID string) (string, error)
	GetOnOffRampWidget(ctx context.Context, walletID string, isDeposit bool) (string, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	CreateWithdrawal(ctx context.Context, walletID, externalTransactionID string) (string, error)

	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
}

type Await func(ctx context.Context, result interface{}) error
