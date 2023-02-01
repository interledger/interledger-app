package ops

import (
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"testing"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/providers/machnet"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	LinkedAccounts() linkedaccounts.Client
	Machnet() machnet.Client
	Transactions() transactions.Client
	Analytics() analytics.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	la  linkedaccounts.Client
	mc  machnet.Client
	tc  transactions.Client
	ac  analytics.Client
}

func (t testBackends) Transactions() transactions.Client {
	return t.tc
}

func (t testBackends) Machnet() machnet.Client {
	return t.mc
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

func NewTestBackends(_ *testing.T, db *sqlx.DB, la linkedaccounts.Client, mc machnet.Client, tc transactions.Client) Backends {
	ac := analytics_client.New(nil, "")
	return &testBackends{db: db, val: validator.New(), la: la, mc: mc, tc: tc, ac: ac}
}
