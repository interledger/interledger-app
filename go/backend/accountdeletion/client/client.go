package client

import (
	"context"

	"gitlab.com/fynbos/backend/accountdeletion"
	"gitlab.com/fynbos/backend/accountdeletion/ops"
)

var _ accountdeletion.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) accountdeletion.Client {
	return &client{b: b}
}

func (c client) Request(ctx context.Context, userID string) error {
	return ops.Request(ctx, c.b, userID)
}

func (c client) GetForUser(ctx context.Context, userID string) (*accountdeletion.Request, error) {
	return ops.GetForUser(ctx, c.b, userID)
}

func (c client) Delete(ctx context.Context, userID string) error {
	return ops.Delete(ctx, c.b, userID)
}
