package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() client.Client
	Transactions() transactions.Client
	Validator() *validator.Validate
	Analytics() analytics.Client
	MX() mx.Client
	Tabapay() tabapay.Client
}

type testBackends struct {
	db *sqlx.DB
	ac analytics.Client
}

func (t testBackends) MX() mx.Client {
	return nil
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
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

func (t testBackends) Tabapay() tabapay.Client {
	return nil
}

func NewTestBackends(db *sqlx.DB) Backends {
	return &testBackends{db: db, ac: analytics_client.New(nil, "")}
}
