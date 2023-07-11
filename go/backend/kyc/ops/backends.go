package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/user"

	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Signup() signup.Client
	Temporal() temporal.Client
	Users() user.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	tp  temporal.Client
	uc  user.Client
	sc  signup.Client
}

func (t testBackends) Users() user.Client {
	return t.uc
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Temporal() temporal.Client {
	return t.tp
}

func (t testBackends) Signup() signup.Client {
	return t.sc
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, tp temporal.Client, uc user.Client, sc signup.Client) Backends {
	return &testBackends{db: db, val: validator.New(), tp: tp, uc: uc, sc: sc}
}
