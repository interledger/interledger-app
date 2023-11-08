package workflows

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/aws"
	aws_mock "gitlab.com/fynbos/backend/aws/client/mock"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/basistheory"
	mock_bt "gitlab.com/fynbos/backend/providers/basistheory/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	mock_client "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	mock_wallets "gitlab.com/fynbos/backend/wallets/client/mock"
)

type Backends interface {
	External() external.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	BasisTheory() basistheory.Client
	DB() *sqlx.DB
	AWS() aws.Client
	Wallets() wallets.Client
}

type InputBackends interface {
	DB() *sqlx.DB
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	BasisTheory() basistheory.Client
	AWS() aws.Client
	Wallets() wallets.Client
}

type backends struct {
	InputBackends
	external external.Client
}

func (ob *backends) External() external.Client {
	return ob.external
}

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *mock_client.MockClient
	Kyc            *kyc_mock.MockClient
	Linkedaccounts *linkedaccount_mock.MockClient
	Basistheory    *mock_bt.MockClient
	AwsCliet       *aws_mock.MockClient
	Wc             *mock_wallets.MockClient
}

func (tb *TestBackends) AWS() aws.Client {
	return tb.AwsCliet
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

func (tb *TestBackends) Wallets() wallets.Client {
	return tb.Wc
}

func NewTestBackends(opts ...func(b *TestBackends)) *TestBackends {
	b := &TestBackends{}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
