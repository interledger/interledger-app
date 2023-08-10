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

func (c client) Create(ctx context.Context, payment payments.Payment) (*payments.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) Update(ctx context.Context, payment payments.Payment) (*payments.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) Confirm(ctx context.Context, paymentID string) (*payments.Payment, error) {
	//TODO implement me
	panic("implement me")
}
