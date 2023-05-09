package platforms

import (
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func NewTestBackends(_ *testing.T, db *sqlx.DB) Backends {
	return &testBackends{db: db, val: validator.New()}
}
