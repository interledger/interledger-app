package ops

import (
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Analytics() analytics.Client
}

type testBackends struct {
	db        *sqlx.DB
	val       *validator.Validate
	analytics analytics.Client
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Analytics() analytics.Client {
	return t.analytics
}

func NewTestBackends(_ *testing.T, db *sqlx.DB) Backends {
	return &testBackends{db: db, val: validator.New(), analytics: analytics_client.New(nil, "")}
}
