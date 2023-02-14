package client

import (
	"context"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
)

var _ authorisation.InternalClient = client{}

type client struct {
	b ops.Backends
}

func (c client) BaseURL() string {
	return ops.BaseURL()
}

func New(b ops.Backends) authorisation.InternalClient {
	return &client{
		b: b,
	}
}

func (c client) AddPublicKey(ctx context.Context, clientURL string, publicKey authorisation.Jwk) error {
	return ops.CreateClientPublicKey(ctx, c.b, clientURL, publicKey)
}
func (c client) Introspect(ctx context.Context, token string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) ListKeys(ctx context.Context, clientURL string) ([]authorisation.Jwk, error) {
	return ops.ListKeys(ctx, c.b, clientURL)
}
