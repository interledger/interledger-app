package ops

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Wallets() wallets.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	wc  wallets.Client
}

func (t testBackends) Wallets() wallets.Client {
	return t.wc
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, wc wallets.Client) Backends {
	return &testBackends{db: db, val: validator.New(), wc: wc}
}
