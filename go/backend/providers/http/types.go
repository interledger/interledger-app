package http

import (
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
