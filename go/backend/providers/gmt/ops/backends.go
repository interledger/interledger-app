package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/identities"

	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/payments"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Temporal() client.Client
	Transactions() transactions.Client
	Validator() *validator.Validate
	Analytics() analytics.Client
	MX() mx.Client
	Tabapay() tabapay.Client
	Keys() keys.Client
	Limits() limits.Client
	Payments() payments.Client
	Wallets() wallets.Client
	Identities() identities.Client
}

type testBackends struct {
	db *sqlx.DB
	ac analytics.Client
	kc keys.Client
}

func (t testBackends) Identities() identities.Client {
	return nil
}

func (t testBackends) Wallets() wallets.Client {
	return nil
}

func (t testBackends) Payments() payments.Client {
	return nil
}

func (t testBackends) Limits() limits.Client {
	return nil
}

func (t testBackends) MX() mx.Client {
	return nil
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) KYC() kyc.Client {
	return nil
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return nil
}

func (t testBackends) Temporal() client.Client {
	return nil
}

func (t testBackends) Transactions() transactions.Client {
	return nil
}

func (t testBackends) Validator() *validator.Validate {
	return nil
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) Tabapay() tabapay.Client {
	return nil
}

func NewTestBackends(t *testing.T, db *sqlx.DB) Backends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	return &testBackends{db: db, ac: analytics_client.New(nil, ""), kc: kc}
}
