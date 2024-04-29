package gatehub

import "context"

type Client interface {
	CreateUser(ctx context.Context, walletID string) (Await, error)
	GetOnboardingWidget(ctx context.Context, walletID string) (string, error)
	GetOnOffRampWidget(ctx context.Context, walletID string, isDeposit bool) (string, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	CreateWithdrawal(ctx context.Context, walletID, externalTransactionID string) (string, error)
}

type Await func(ctx context.Context, result interface{}) error
