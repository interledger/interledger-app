package client

import (
	"context"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/dynamicforms"
	"gitlab.com/fynbos/backend/dynamicforms/ops"
	"io"
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

func (c client) ListFormCounts(ctx context.Context, page db.Pagination) ([]dynamicforms.FormCount, error) {
	return ops.ListFormCounts(ctx, c.b, page)
}

func (c client) ExportFormResults(ctx context.Context, formID string, writer io.Writer) error {
	return ops.ExportFormResults(ctx, c.b, formID, writer)
}
