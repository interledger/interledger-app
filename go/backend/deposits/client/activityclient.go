package client

import (
	"context"

	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/deposits/ops"
)

type activityClient struct {
	b ops.Backends
}

func (a activityClient) CreatePendingDeposit(ctx context.Context, depositId string) (string, error) {
	return ops.CreatePendingDeposit(ctx, a.b, depositId)
}

func (a activityClient) ProcessNoopDeposit(ctx context.Context, depositId string) error {
	return ops.ProcessNoopDeposit(ctx, a.b, depositId)
}

func (a activityClient) VoidPendingDeposit(ctx context.Context, trxId string) error {
	return ops.VoidPendingDeposit(ctx, a.b, trxId)
}

func (a activityClient) PostPendingDeposit(ctx context.Context, trxId string) error {
	return ops.PostPendingDeposit(ctx, a.b, trxId)
}

func (a activityClient) SetDepositState(ctx context.Context, depositId string, state deposits.State) error {
	return ops.SetDepositState(ctx, a.b, depositId, state)
}

func MakeActivity(b Backends) deposits.ActivityClient {
	return activityClient{b: b}
}
