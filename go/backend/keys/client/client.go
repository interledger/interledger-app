package client

import (
	"context"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/keys/ops"
)

var _ keys.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) keys.Client {
	return &client{
		b: b,
	}
}

func (c client) ProvisionPrivateKey(ctx context.Context, walletID string) error {
	return ops.GeneratePrivateKey(ctx, c.b, walletID)
}

func (c client) AddPublicKey(ctx context.Context, walletID string, publicKeyBase64 string, name string) error {
	return ops.AddPublicKey(ctx, c.b, walletID, publicKeyBase64, name)
}
