package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Analytics() analytics.Client
	Temporal() temporal.Client
	Users() user.Client
}

type testBackends struct {
	db        *sqlx.DB
	val       *validator.Validate
	analytics analytics.Client
	tp        temporal.Client
	uc        user.Client
}

func (t testBackends) Users() user.Client {
	return t.uc
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

func (t testBackends) Temporal() temporal.Client {
	return t.tp
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, tp temporal.Client, uc user.Client) Backends {
	return &testBackends{db: db, val: validator.New(), analytics: analytics_client.New(nil, ""), tp: tp, uc: uc}
}
