package ops_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
)


func NewTestBackends(t *testing.T, Db *sqlx.DB) *TestBackends {

	return &TestBackends{
		Db: Db,
	}
}

type TestBackends struct {
	Db *sqlx.DB
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}
