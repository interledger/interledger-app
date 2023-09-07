package dynamicforms

import (
	"context"
	"gitlab.com/fynbos/backend/db"
	"io"
)

type Client interface {
	Create(ctx context.Context, args *CreateFormArgs) (*Form, error)
	ListFormCounts(ctx context.Context, page db.Pagination) ([]FormCount, error)
	ExportFormResults(ctx context.Context, formID string, writer io.Writer) error
}
