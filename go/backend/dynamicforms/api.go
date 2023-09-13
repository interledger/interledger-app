package dynamicforms

import (
	"context"
	"gitlab.com/fynbos/backend/db"
	"io"
)

type Client interface {
	Submit(ctx context.Context, args *SubmitArgs) (*Submission, error)
	ListSubmissionCounts(ctx context.Context, page db.Pagination) ([]SubmissionCount, error)
	ExportSubmissions(ctx context.Context, formID string, writer io.Writer) error
	ListSubmissions(ctx context.Context, formID string) ([]Submission, error)
	GetSubmission(ctx context.Context, id string) (*Submission, error)
}
