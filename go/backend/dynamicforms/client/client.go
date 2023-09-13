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

func (c client) Submit(ctx context.Context, args *dynamicforms.SubmitArgs) (*dynamicforms.Submission, error) {
	return ops.SubmitForm(ctx, c.b, args)
}

func (c client) ListSubmissionCounts(ctx context.Context, page db.Pagination) ([]dynamicforms.SubmissionCount, error) {
	return ops.ListSubmissionCount(ctx, c.b, page)
}

func (c client) ExportSubmissions(ctx context.Context, formID string, writer io.Writer) error {
	return ops.ExportSubmissions(ctx, c.b, formID, writer)
}

func (c client) ListSubmissions(ctx context.Context, formID string) ([]dynamicforms.Submission, error) {
	return ops.ListSubmissions(ctx, c.b, formID)
}

func (c client) GetSubmission(ctx context.Context, id string) (*dynamicforms.Submission, error) {
	return ops.GetSubmission(ctx, c.b, id)
}
