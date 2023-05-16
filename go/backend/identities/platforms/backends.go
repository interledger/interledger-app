package platforms

import (
	"gitlab.com/fynbos/backend/openpayments"
	"testing"

	temporal "go.temporal.io/sdk/client"

	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/twitter"

	"github.com/go-playground/validator/v10"

	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Keys() keys.Client
	Analytics() analytics.Client
	Temporal() temporal.Client
	OpenPayments() openpayments.Client
}

type testBackends struct {
	db      *sqlx.DB
	val     *validator.Validate
	an      analytics.Client
	keys    keys.Client
	twitter twitter.Client
	op      openpayments.Client
}

func (t testBackends) Temporal() temporal.Client {
	//TODO implement me
	panic("implement me")
}

func (t testBackends) Analytics() analytics.Client {
	return t.an
}

func (t testBackends) Twitter() twitter.Client {
	return t.twitter
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) Keys() keys.Client {
	return t.keys
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) OpenPayments() openpayments.Client {
	return t.op
}

func NewTestBackends(_ *testing.T, db *sqlx.DB) Backends {
	return &testBackends{db: db, val: validator.New()}
}
