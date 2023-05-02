package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/basistheory/ops"
)

var _ basistheory.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) basistheory.Client {
	return &client{
		b: b,
	}
}

func (c client) CreateCard(ctx context.Context, tokenID, walletID string) (*basistheory.Card, error) {
	return ops.CreateCard(ctx, c.b, tokenID, walletID)
}

func (c client) GetCard(ctx context.Context, id string) (*basistheory.Card, error) {
	return ops.GetCard(ctx, c.b, id)
}
