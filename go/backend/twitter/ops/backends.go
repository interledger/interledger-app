package ops

import (
	"github.com/jmoiron/sqlx"
)

type (
	Backends interface {
		DB() *sqlx.DB
	}

	TestBackends struct {
		Db *sqlx.DB
	}
)

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func NewTestBackends(opts ...func(*TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
