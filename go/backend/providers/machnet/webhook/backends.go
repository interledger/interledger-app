package webhook

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	Machnet() machnet.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() client.Client
}

type opsBackends struct {
	Backends
}

func (b opsBackends) External() external.Client {
	return b.Machnet().External()
}

func NewTestBackends(t *testing.T) testBackends {
	ctrl := gomock.NewController(t)
	return testBackends{
		db:             test_utils.MigrateCockroachDB(t, context.Background()),
		external:       external_client.New(),
		linkedaccounts: linkedaccounts_mock.NewMockClient(ctrl),
		temporal:       &mocks.Client{},
	}
}

type testBackends struct {
	db             *sqlx.DB
	external       *external_client.Client
	linkedaccounts *linkedaccounts_mock.MockClient
	users          user.Client
	kycImpl        kyc.Client
	temporal       *mocks.Client
}

func (b testBackends) Machnet() machnet.Client {
	return nil
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

func (b testBackends) External() external.Client {
	return b.external
}

func (b testBackends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}

func (b testBackends) Temporal() client.Client {
	return b.temporal
}
