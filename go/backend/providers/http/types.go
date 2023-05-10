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
	Provider string
	Context  string
}

type contextKey struct {
	name string
}

var ContextKey = &contextKey{"httplog_metadata"}

const insertFields = "provider, context, request_body, request_path, response_body, response_status"

// const fields = "id, provider, context, request_body, request_path, response_body, response_status, created_at"

type LogRecord struct {
	ID             string    `db:"id"`
	Provider       string    `db:"provider"`
	Context        string    `db:"context"`
	RequestPath    string    `db:"request_path"`
	RequestBody    string    `db:"request_body"`
	ResponseBody   string    `db:"response_body"`
	ResponseStatus string    `db:"response_status"`
	CreatedAt      time.Time `db:"created_at"`
}

type Redact func(ctx context.Context, req []byte) ([]byte, error)
