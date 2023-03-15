package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	External() external.Service
	LinkedAccounts() linkedaccounts.Client
	Temporal() client.Client
	Transactions() transactions.Client
	Validator() *validator.Validate
}

type testBackends struct {
	db *sqlx.DB
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) KYC() kyc.Client {
	return nil
}

func (t testBackends) External() external.Service {
	return nil
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return nil
}

func (t testBackends) Temporal() client.Client {
	return nil
}

func (t testBackends) Transactions() transactions.Client {
	return nil
}

func (t testBackends) Validator() *validator.Validate {
	return nil
}

func NewTestBackends(db *sqlx.DB) Backends {
	return &testBackends{db: db}
}
