package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/providers/machnet"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	LinkedAccounts() linkedaccounts.Client
	Machnet() machnet.Client
	Transactions() transactions.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	la  linkedaccounts.Client
	mc  machnet.Client
	tc  transactions.Client
}

func (t testBackends) Transactions() transactions.Client {
	return nil
}

func (t testBackends) Machnet() machnet.Client {
	return t.mc
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, mc machnet.Client) Backends {
	return &testBackends{db: db, val: validator.New(), la: la, mc: mc}
}
