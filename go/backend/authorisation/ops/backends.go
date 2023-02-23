package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/openpayments"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	OpenPayments() openpayments.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	op  openpayments.Client
}

func (t testBackends) OpenPayments() openpayments.Client {
	return t.op
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, op openpayments.Client) Backends {
	return &testBackends{db: db, val: validator.New(), op: op}
}
