package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	external_client "gitlab.com/fynbos/backend/providers/mx/external/client"
	"gitlab.com/fynbos/backend/providers/mx/ops"
)

type Backends interface{}

var _ mx.Client = &Client{}

type opsBackends struct {
	external external.Client
}

func (b *opsBackends) External() external.Client {
	return b.external
}

func New(clientID, apiKey string) *Client {
	return &Client{
		b: &opsBackends{
			external: external_client.New(clientID, apiKey),
		},
	}
}

type Client struct {
	b ops.Backends
}

func (c *Client) GetWidget(ctx context.Context, walletID string) (string, error) {
	return ops.GetWidget(ctx, c.b, walletID)
}
