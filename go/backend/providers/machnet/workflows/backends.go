package workflows

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
}

type opsBackends struct {
	Backends
	external external.Client
}

func (b opsBackends) External() external.Client {
	return b.external
}

type testBackends struct {
	db      *sqlx.DB
	users   user.Client
	kycImpl *kyc_mock.MockClient
	linked  *linkedaccounts_mock.MockClient
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
