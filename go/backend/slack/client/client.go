package client

import (
	"context"

	"gitlab.com/fynbos/backend/payments"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/slack/external"

	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/slack/ops"
)

type Backends interface {
	DB() *sqlx.DB
	Payments() payments.Client
}

var _ ops.Backends = opsBackends{}

type opsBackends struct {
	Backends
	slackExt external.Client
}

func (ob opsBackends) External() external.Client {
	return ob.slackExt
}

var _ slack.Client = &client{}

type client struct {
	b ops.Backends
}

func New(b Backends, cfg external.Config) (slack.Client, error) {
	se, err := external.New(cfg)
	if err != nil {
		return nil, err
	}

	return &client{b: &opsBackends{Backends: b, slackExt: se}}, nil
}

func (c *client) CreateAuthURL(ctx context.Context, walletID string) (string, error) {
	return ops.CreateAuthURL(ctx, c.b, walletID)
}

func (c *client) CreateConnection(ctx context.Context, args slack.CreateConnectionArgs) (*slack.Connection, error) {
	return ops.CreateConnection(ctx, c.b, args)
}
