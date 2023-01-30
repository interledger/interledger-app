package client

import (
	"context"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
)

var _ authorisation.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) authorisation.Client {
	return &client{
		b: b,
	}
}

func (c client) AddPublicKey(ctx context.Context, walletID string, key string) error {
	//TODO implement me
	panic("implement me")
}
func (c client) Introspect(ctx context.Context, token string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) ListKeys(ctx context.Context, paymentPointer string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}
