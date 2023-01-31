package server

import (
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"testing"

	"gitlab.com/fynbos/backend/authorisation"

	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/providers/machnet"

	"gitlab.com/fynbos/backend/email"

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
	KYC() kyc.Client
	Transactions() transactions.Client
	Authorisation() authorisation.Client
	Analytics() analytics.Client
}

type testBackends struct {
	db   *sqlx.DB
	val  *validator.Validate
	la   linkedaccounts.Client
	temp temporal.Client
	em   email.Client
	mc   machnet.Client
	tr   transactions.Client
	ac   analytics.Client
}

func (t testBackends) Authorisation() authorisation.Client {
	return nil
}

func (t testBackends) KYC() kyc.Client {
	return nil
}

func (t testBackends) Machnet() machnet.Client {
	return t.mc
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

func (t testBackends) Transactions() transactions.Client {
	return t.tr
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, temp temporal.Client, mc machnet.Client, tr transactions.Client) Backends {
	ac := analytics_client.New(nil, "")
	return &testBackends{db: db, val: validator.New(), la: la, temp: temp, mc: mc, tr: tr, ac: ac}
}
