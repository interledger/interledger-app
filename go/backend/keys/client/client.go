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

func (c client) AddPublicKey(ctx context.Context, walletID, publicKeyBase64 string, name, keyID string) (*keys.Key, error) {
	return ops.AddPublicKey(ctx, c.b, walletID, publicKeyBase64, name, keyID)
}

func (c client) DeletePublicKey(ctx context.Context, id string) error {
	return ops.DeletePublicKey(ctx, c.b, id)
}

func (c client) List(ctx context.Context, walletID string) ([]keys.Key, error) {
	return ops.ListKeys(ctx, c.b, walletID)
}

func (c client) GetPublicKey(ctx context.Context, id string, walletID string) (*keys.Key, error) {
	return ops.GetPublicKey(ctx, c.b, id, walletID)
}

func (c client) RemoveCustodialKeysForWallet(ctx context.Context, walletID string) error {
	return ops.RemoveCustodialKeysForWallet(ctx, c.b, walletID)
}
