package workflows

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/basistheory"
	mock_bt "gitlab.com/fynbos/backend/providers/basistheory/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	mock_client "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	BasisTheory() basistheory.Client
}

type InputBackends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	BasisTheory() basistheory.Client
}

type backends struct {
	external external.Client
	b        InputBackends
}

func (ob *backends) External() external.Client {
	return ob.external
}

func (ob *backends) KYC() kyc.Client {
	return ob.b.KYC()
}

func (ob *backends) LinkedAccounts() linkedaccounts.Client {
	return ob.b.LinkedAccounts()
}

func (ob *backends) BasisTheory() basistheory.Client {
	return ob.b.BasisTheory()
}

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *mock_client.MockClient
	Kyc            *kyc_mock.MockClient
	Linkedaccounts *linkedaccount_mock.MockClient
	Basistheory    *mock_bt.MockClient
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func (tb *TestBackends) BasisTheory() basistheory.Client {
	return tb.Basistheory
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}

func (tb *TestBackends) KYC() kyc.Client {
	return tb.Kyc
}

func (tb *TestBackends) LinkedAccounts() linkedaccounts.Client {
	return tb.Linkedaccounts
}

func NewTestBackends(opts ...func(b *TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
