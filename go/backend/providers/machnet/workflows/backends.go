package workflows

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock "gitlab.com/fynbos/backend/providers/machnet/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Machnet() machnet.Client
	Temporal() client.Client
	Transactions() transactions.Client
}

type opsBackends struct {
	Backends
	external external.Client
}

func (b opsBackends) External() external.Client {
	return b.external
}

type testBackends struct {
	db       *sqlx.DB
	users    user.Client
	kycImpl  *kyc_mock.MockClient
	linked   *linkedaccounts_mock.MockClient
	machnet  *machnet_mock.MockClient
	temporal *mocks.Client
	tx       transactions.Client
}

func (b testBackends) LinkedAccounts() linkedaccounts.Client {
	return b.linked
}

func (b testBackends) Users() user.Client {
	return b.users
}

func (b testBackends) KYC() kyc.Client {
	return b.kycImpl
}

func (b testBackends) DB() *sqlx.DB {
	return b.db
}

func (b testBackends) Validator() *validator.Validate {
	return validator.New()
}

func (b testBackends) Machnet() machnet.Client {
	return b.machnet
}

func (b testBackends) Temporal() client.Client {
	return b.temporal
}

func (b testBackends) Transactions() transactions.Client {
	return b.tx
}
