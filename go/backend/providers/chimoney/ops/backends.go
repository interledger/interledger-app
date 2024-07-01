package ops

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	LinkedAccounts() linkedaccounts.Client
	Users() user.Client
	Temporal() temporal.Client
	Wallets() wallets.Client
	Pacioli() pacioli.Client
	KYC() kyc.Client
	Transactions() transactions.Client
	Email() email.Client
}

type TestBackends struct {
	DBC *sqlx.DB
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	return nil
}

func (t TestBackends) Users() user.Client {
	return nil
}

func (t TestBackends) Temporal() temporal.Client {
	return nil
}

func (t TestBackends) Wallets() wallets.Client {
	return nil
}

func (t TestBackends) Pacioli() pacioli.Client {
	return nil
}

func (t TestBackends) KYC() kyc.Client {
	return nil
}

func (t TestBackends) DB() *sqlx.DB {
	return t.DBC
}

func (t TestBackends) Transactions() transactions.Client {
	return nil
}

func (t TestBackends) Email() email.Client {
	return nil
}

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
