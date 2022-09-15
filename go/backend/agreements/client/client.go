package client

import (
	"context"

	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/agreements/ops"
)

var _ agreements.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) agreements.Client {
	return &client{
		b: b,
	}
}

func (c client) Sign(ctx context.Context, args *agreements.SignArgs) error {
	return ops.Sign(ctx, c.b, args)
}

func (c client) GetSignatures(ctx context.Context, identityID string) ([]agreements.Signature, error) {
	return ops.GetSignatures(ctx, c.b, identityID)
}

func (c client) Get(ctx context.Context, id string) (*agreements.Agreement, error) {
	return ops.Get(ctx, c.b, id)
}
