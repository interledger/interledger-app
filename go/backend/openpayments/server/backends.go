package server

import (
	"gitlab.com/fynbos/backend/contacts"
	"testing"

	"gitlab.com/fynbos/backend/limits"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	Users() user.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	KYC() kyc.Client
	Transactions() transactions.Client
	Authorisation() authorisation.InternalClient
	Analytics() analytics.Client
	Limits() limits.Client
	Contacts() contacts.Client
}

type testBackends struct {
	db   *sqlx.DB
	val  *validator.Validate
	la   linkedaccounts.Client
	temp temporal.Client
	em   email.Client
	tr   transactions.Client
	ac   analytics.Client
	auth authorisation.InternalClient
	lmt  limits.Client
	cc   contacts.Client
}

func (t testBackends) Authorisation() authorisation.InternalClient {
	return t.auth
}

func (t testBackends) KYC() kyc.Client {
	return nil
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

func (t testBackends) Limits() limits.Client {
	return t.lmt
}

func (t testBackends) Contacts() contacts.Client {
	return t.cc
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, temp temporal.Client, tr transactions.Client, auth authorisation.InternalClient, lmt limits.Client, cc contacts.Client) Backends {
	ac := analytics_client.New(nil, "")
	return &testBackends{db: db, val: validator.New(), la: la, temp: temp, tr: tr, ac: ac, auth: auth, lmt: lmt, cc: cc}
}
