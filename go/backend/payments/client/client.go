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

func New(b ops.Backends) payments.Client {
	return &client{
		b: b,
	}
}

func (c client) Lookup(ctx context.Context, id string) (*payments.Payment, error) {
	return ops.Lookup(ctx, c.b, id)
}

func (c client) Create(ctx context.Context, args payments.CreateArgs) (*payments.Payment, error) {
	return ops.Create(ctx, c.b, args)
}

func (c client) Update(ctx context.Context, args payments.UpdateArgs) (*payments.Payment, error) {
	return ops.Update(ctx, c.b, args)
}

func (c client) Confirm(ctx context.Context, id string) (*payments.Payment, []payments.RequiredActionType, error) {
	return ops.Confirm(ctx, c.b, id)
}

func (c client) UpdateReceiver(ctx context.Context, id string, identity payments.Identity) error {
	return ops.UpdateReceiver(ctx, c.b, id, identity)
}

func (c client) SignalIdentityCreated(ctx context.Context, identifier string) error {
	return ops.SignalIdentityCreated(ctx, c.b, identifier)
}

func (c client) SignalAccountLinked(ctx context.Context, walletID string) error {
	return ops.SignalAccountLinked(ctx, c.b, walletID)
}

func (c client) AdminListAwaitingSignal(ctx context.Context) ([]payments.Payment, error) {
	return ops.ListAwaitingSignal(ctx, c.b)
}
