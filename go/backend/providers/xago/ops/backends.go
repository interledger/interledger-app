package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/pacioli"

	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/xago/external"
	external_mock "gitlab.com/fynbos/backend/providers/xago/external/mock"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Payments() payments.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	Pacioli() pacioli.Client
	Transactions() transactions.Client
}

type ActivityBackends interface {
	DB() *sqlx.DB
	Payments() payments.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	Pacioli() pacioli.Client
	Transactions() transactions.Client
}

type TestBackends struct {
	DBC  *sqlx.DB
	Extr *external_mock.MockClient
	La   *linkedaccounts_mock.MockClient
}

func (t TestBackends) Transactions() transactions.Client {
	return nil
}

func (t TestBackends) Pacioli() pacioli.Client {
	return nil
}

func (t TestBackends) DB() *sqlx.DB {
	return t.DBC
}

func (t TestBackends) External() external.Client {
	return t.Extr
}

func (t TestBackends) Payments() payments.Client {
	return nil
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	return t.La
}

func (t TestBackends) Wallets() wallets.Client {
	return nil
}

func (t TestBackends) Users() user.Client {
	return nil
}

func (t TestBackends) KYC() kyc.Client {
	return nil
}

func (t TestBackends) Temporal() temporal.Client {
	return nil
}

func NewTestBackends(_ *testing.T, opts ...func(*TestBackends)) Backends {
	b := &TestBackends{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
