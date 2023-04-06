package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	LinkedAccounts() linkedaccounts.Client
	Transactions() transactions.Client
	Analytics() analytics.Client
	Tabapay() tabapay.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	la  linkedaccounts.Client
	tc  transactions.Client
	ac  analytics.Client
	tbc tabapay.Client
}

func (t testBackends) Transactions() transactions.Client {
	return t.tc
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) Tabapay() tabapay.Client {
	return t.tbc
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, tc transactions.Client) Backends {
	ac := analytics_client.New(nil, "")
	return &testBackends{db: db, val: validator.New(), la: la, tc: tc, ac: ac}
}
