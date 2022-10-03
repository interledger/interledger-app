package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	inmemory_external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
)

// const (
// 	sandboxUrl = "https://v4sandbox.machpay.com/v4"
// 	prodUrl    = "https://machpay.com"
// )

type Backends interface {
	DB() *sqlx.DB
}

type opsBackends struct {
	Backends
	external external.Client
}

func (b opsBackends) DB() *sqlx.DB {
	return b.Backends.DB()
}

func (b opsBackends) External() external.Client {
	return b.external
}

func New(b Backends) machnet.Client {
	// TODO: http client

	opsBackends := opsBackends{
		Backends: b,
		external: inmemory_external_client.New(),
	}

	return &client{b: opsBackends}
}

type client struct {
	b ops.Backends
}

func (c client) GetUser(ctx context.Context, walletID string) (*machnet.User, error) {
	return ops.GetUser(ctx, c.b, walletID)
}

func (c client) CreateUser(ctx context.Context, args machnet.CreateArgs) (*machnet.User, error) {
	return ops.CreateUser(ctx, c.b, args)
}

func (c client) GetWidgetToken(ctx context.Context, walletID string) (*machnet.WidgetToken, error) {
	return ops.GetWidgetToken(ctx, c.b, walletID)
}
