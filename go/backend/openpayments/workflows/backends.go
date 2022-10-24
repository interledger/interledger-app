package workflows

import (
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
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	t   temporal.Client
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	//TODO implement me
	panic("implement me")
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

func NewTestBackends(_ *testing.T, db *sqlx.DB, temp temporal.Client) Backends {
	return &testBackends{db: db, val: validator.New()}
}
