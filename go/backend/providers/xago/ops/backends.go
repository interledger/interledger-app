package ops

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/email"

	"github.com/interledger/interledger-app/go/backend/transactions"

	"github.com/interledger/interledger-app/go/pacioli"

	linkedaccounts_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"

	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/xago/external"
	external_mock "github.com/interledger/interledger-app/go/backend/providers/xago/external/mock"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/jmoiron/sqlx"
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
	Email() email.Client
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
	Email() email.Client
}

type TestBackends struct {
	DBC  *sqlx.DB
	Extr *external_mock.MockClient
	La   *linkedaccounts_mock.MockClient
}

func (t TestBackends) Email() email.Client {
	return nil
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
