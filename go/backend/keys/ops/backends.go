package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/rafiki"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Rafiki() rafiki.Client
}

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	rafiki rafiki.Client
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Rafiki() rafiki.Client {
	return t.rafiki
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, rc rafiki.Client) Backends {
	return &testBackends{db: db, val: validator.New(), rafiki: rc}
}
