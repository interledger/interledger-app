package workflows

import (
	"testing"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/providers/machnet"

	"gitlab.com/fynbos/backend/linkedaccounts"

	temporal "go.temporal.io/sdk/client"

	"gitlab.com/fynbos/backend/user"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	Users() user.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	Machnet() machnet.Client
	Transactions() transactions.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	t   temporal.Client
	la  linkedaccounts.Client
	em  email.Client
	tx  transactions.Client
}

func (t testBackends) Transactions() transactions.Client {
	return t.tx
}

func (t testBackends) Machnet() machnet.Client {
	return nil
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Email() email.Client {
	return t.em
}

func (t testBackends) Temporal() temporal.Client {
	return t.t
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, temp temporal.Client, la linkedaccounts.Client, tx transactions.Client) Backends {
	return &testBackends{db: db, val: validator.New(), t: temp, la: la, tx: tx}
}
