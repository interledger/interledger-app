package ops_test

import "github.com/jmoiron/sqlx"

type TestBackends struct {
	Db *sqlx.DB
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}
