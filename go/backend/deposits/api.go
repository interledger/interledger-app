package deposits

import (
	"context"
)

type Client interface {
	InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*Deposit, error)
	Get(ctx context.Context, id string) (*Deposit, error)
	SetState(ctx context.Context, id string, state State) error
}

type ActivityClient interface {
	CreatePendingDeposit(ctx context.Context, depositId string) (string, error)
	ProcessNoopDeposit(ctx context.Context, depositId string) error
	VoidPendingDeposit(ctx context.Context, trxId string) error
	PostPendingDeposit(ctx context.Context, trxId string) error
	SetDepositState(ctx context.Context, depositId string, state State) error
}
