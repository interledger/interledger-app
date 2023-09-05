package client

import (
	"context"
	"gitlab.com/fynbos/backend/dynamicforms"
	"gitlab.com/fynbos/backend/dynamicforms/ops"
)

var _ dynamicforms.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) dynamicforms.Client {
	return &client{
		b: b,
	}
}

func (c client) Create(ctx context.Context, args *dynamicforms.CreateFormArgs) (*dynamicforms.Form, error) {
	return ops.CreateForm(ctx, c.b, args)
}
