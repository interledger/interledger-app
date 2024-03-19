package client

import (
	"context"
	"os"

	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	ops "gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/user"
)

var _ gatehub.Client = Client{}

type Client struct {
	b  ops.Backends
	ec external.Client
}

type Backends interface {
	Users() user.Client
}

func New(b Backends) *Client {
	ec := external.NewClient(os.Getenv("GATEHUB_APP_ID"), os.Getenv("GATEHUB_SECRET"))

	return &Client{
		b:  b,
		ec: ec,
	}
}

func (c Client) CreateUser(ctx context.Context, walletID string) error {
	return ops.CreateUser(ctx, c.b, c.ec, walletID)
}
