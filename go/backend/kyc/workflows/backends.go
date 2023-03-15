package workflows

import (
	"gitlab.com/fynbos/backend/kyc"
	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	Temporal() temporal.Client
	KYC() kyc.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	t   temporal.Client
	kyc kyc.Client
}

func (t testBackends) Temporal() temporal.Client {
	return t.t
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

type testBackendArgs struct {
	db  *sqlx.DB
	tp  temporal.Client
	kyc kyc.Client
}

func NewTestBackends(args testBackendArgs) Backends {
	return &testBackends{db: args.db, val: validator.New(), t: args.tp, kyc: args.kyc}
}
