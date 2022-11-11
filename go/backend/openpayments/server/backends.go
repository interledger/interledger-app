package server

import (
	"gitlab.com/fynbos/backend/email"
	"testing"

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
}

type testBackends struct {
	db   *sqlx.DB
	val  *validator.Validate
	la   linkedaccounts.Client
	temp temporal.Client
	em   email.Client
}

func (t testBackends) Temporal() temporal.Client {
	return t.temp
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Email() email.Client {
	return t.em
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

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, temp temporal.Client) Backends {
	return &testBackends{db: db, val: validator.New(), la: la, temp: temp}
}
