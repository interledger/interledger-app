package http

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
}

type Metadata struct {
	Provider string `json:"provider"`
	Context  string `json:"context"`
	Method   string `json:"method"`
}

func MetaForContext(ctx context.Context) (*Metadata, bool) {
	m, ok := ctx.Value(ContextKey).(*Metadata)
	return m, ok
}

type contextKey struct {
	name string
}

var ContextKey = &contextKey{"httplog_metadata"}

const insertFields = "provider, context, request_body, request_path, response_body, response_status, method"

const Fields = "id, provider, context, request_body, request_path, response_body, response_status, created_at"

type LogRecord struct {
	ID             string    `db:"id"`
	Provider       string    `db:"provider"`
	Context        string    `db:"context"`
	Method         string    `db:"method"`
	RequestPath    string    `db:"request_path"`
	RequestBody    string    `db:"request_body"`
	ResponseBody   string    `db:"response_body"`
	ResponseStatus string    `db:"response_status"`
	CreatedAt      time.Time `db:"created_at"`
}

type Redact func(ctx context.Context, req []byte) ([]byte, error)
