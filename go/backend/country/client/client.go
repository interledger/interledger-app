package client

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/country/ops"
)

var _ country.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) country.Client {
	return &client{
		b: b,
	}
}

func (c client) Get(ctx context.Context, id string) (*country.Country, error) {
	return ops.Get(ctx, c.b, id)
}

func (c client) GetByAlpha2(ctx context.Context, code string) (*country.Country, error) {
	return ops.GetByAlpha2(ctx, c.b, code)
}

func (c client) GetAll(ctx context.Context) ([]country.Country, error) {
	return ops.GetAll(ctx, c.b)
}
