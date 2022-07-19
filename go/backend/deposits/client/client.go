package client

import (
	"context"

	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/deposits/ops"
)

type client struct {
	b ops.Backends
}

func Make(b Backends) deposits.Client {
	// TODO Add any service spesific requirements that the rest of the application does not need. i.e. Third party cleints only used by the service
	return client{b: b}
}

func (c client) InitiateDeposit(ctx context.Context, args *deposits.InitiateDepositArgs) (*deposits.Deposit, error) {
	return ops.InitiateDeposit(ctx, c.b, args)
}

func (c client) Get(ctx context.Context, id string) (*deposits.Deposit, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) SetState(ctx context.Context, id string, state deposits.State) error {
	return ops.SetState(ctx, c.b, id, state)
}
