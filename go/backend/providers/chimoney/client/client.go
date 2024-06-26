package client

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ chimoney.Client = &Client{}

type Client struct {
	b        ops.Backends
	external external.Client
}

func New(b ops.Backends) chimoney.Client {
	return &Client{
		b: b,
		external: external.New(
			&http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
				),
			},
		),
	}
}

func (c *Client) CreateWallet(ctx context.Context, walletID string) (chimoney.Await, error) {
	return ops.CreateWallet(ctx, c.b, walletID)
}

func (c *Client) AddInterlocEmail(ctx context.Context, walletID, email string) (string, error) {
	return ops.UpsertInterlocEmail(ctx, c.b, walletID, email)
}

func (c *Client) CreateDepositLink(ctx context.Context, walletID string, amt currency.Amount) (string, error) {
	return ops.CreateDepositLink(ctx, c.b, c.external, walletID, amt)
}
