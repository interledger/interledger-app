package ops

import (
	"testing"

	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() temporal.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
}

func (t testBackends) Temporal() temporal.Client {
	//TODO implement me
	panic("implement me")
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
