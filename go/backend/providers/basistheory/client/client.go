package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
	external_client "gitlab.com/fynbos/backend/providers/basistheory/external/client"
	"gitlab.com/fynbos/backend/providers/basistheory/ops"
)

var _ basistheory.Client = client{}

type Backends interface {
	DB() *sqlx.DB
}

type opsBackends struct {
	b        Backends
	external external.Client
}

func (ob *opsBackends) DB() *sqlx.DB {
	return ob.b.DB()
}

func (ob *opsBackends) External() external.Client {
	return ob.external
}

type client struct {
	b ops.Backends
}

func New(apiKey string, b Backends) basistheory.Client {
	return &client{
		b: &opsBackends{
			b:        b,
			external: external_client.New(apiKey),
		},
	}
}

func (c client) CreateCard(ctx context.Context, args basistheory.CreateCardArgs) (*basistheory.Card, error) {
	return ops.CreateCard(ctx, c.b, args)
}

func (c client) GetCard(ctx context.Context, id string) (*basistheory.Card, error) {
	return ops.GetCard(ctx, c.b, id)
}

func (c client) CreateCardToken(ctx context.Context, args basistheory.CreateCardTokenArgs) (string, error) {
	token, err := c.b.External().CreateCardToken(ctx, args)
	if err != nil {
		return "", err
	}

	return token.GetId(), nil
}

func (c client) UpdateCard(ctx context.Context, args basistheory.UpdateCardArgs) (*basistheory.Card, error) {
	return ops.UpdateCard(ctx, c.b, args)
}

func (c client) ListCards(ctx context.Context) ([]basistheory.Card, error) {
	return ops.ListCards(ctx, c.b, 1000)
}
