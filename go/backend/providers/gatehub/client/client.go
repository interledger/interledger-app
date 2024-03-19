package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/gatehub"
	ops "gitlab.com/fynbos/backend/providers/gatehub/ops"
)

var _ gatehub.Client = Client{}

type Client struct {
	b ops.Backends
}

func New(b ops.Backends) *Client {
	return &Client{b}
}

func (c Client) CreateUser(ctx context.Context, walletID string) (gatehub.Await, error) {
	return ops.CreateUser(ctx, c.b, walletID)
}
