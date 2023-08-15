package client

import (
	"context"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/ops"
)

var _ payments.Client = client{}

type client struct {
	b ops.Backends
}

func (c client) Lookup(ctx context.Context, id string) (*payments.Payment, error) {
	return ops.Lookup(ctx, c.b, id)
}

func (c client) Create(ctx context.Context, args payments.CreateArgs) (*payments.Payment, error) {
	return ops.Create(ctx, c.b, args)
}

func (c client) Update(ctx context.Context, payment payments.Payment) (*payments.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) Confirm(ctx context.Context, id string) (*payments.Payment, []payments.RequiredActionType, error) {
	return ops.Confirm(ctx, c.b, id)
}
