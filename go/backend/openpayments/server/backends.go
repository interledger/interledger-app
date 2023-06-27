package server

import (
	"testing"

	"gitlab.com/fynbos/backend/twilio"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/identities"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/limits"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	Users() user.Client
	Wallets() wallets.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	KYC() kyc.Client
	Transactions() transactions.Client
	Authorisation() authorisation.InternalClient
	Analytics() analytics.Client
	Limits() limits.Client
	Contacts() contacts.Client
	Tabapay() tabapay.Client
	Keys() keys.Client
	Identities() identities.Client
	Features() features.Client
	Twilio() twilio.Service
}

type testBackends struct {
	db   *sqlx.DB
	val  *validator.Validate
	la   linkedaccounts.Client
	temp temporal.Client
	em   email.Client
	tr   transactions.Client
	ac   analytics.Client
	auth authorisation.InternalClient
	lmt  limits.Client
	cc   contacts.Client
	tbc  tabapay.Client
	keys keys.Client
	ids  identities.Client
	us   user.Client
	fc   features.Client
	wc   wallets.Client
}

func (t *testBackends) Wallets() wallets.Client {
	return t.wc
}

func (t *testBackends) Twilio() twilio.Service {
	return nil
}

func (t *testBackends) Keys() keys.Client {
	return t.keys
}

func (t *testBackends) Authorisation() authorisation.InternalClient {
	return t.auth
}

func (t *testBackends) KYC() kyc.Client {
	return nil
}

func (t *testBackends) Temporal() temporal.Client {
	return t.temp
}

func (t *testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t *testBackends) Email() email.Client {
	return t.em
}

func (t *testBackends) Users() user.Client {
	return t.us
}

func (t *testBackends) Validator() *validator.Validate {
	return t.val
}

func (t *testBackends) DB() *sqlx.DB {
	return t.db
}

func (t *testBackends) Transactions() transactions.Client {
	return t.tr
}

func (t *testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t *testBackends) Limits() limits.Client {
	return t.lmt
}

func (t *testBackends) Contacts() contacts.Client {
	return t.cc
}

func (t *testBackends) Tabapay() tabapay.Client {
	return t.tbc
}

func (t *testBackends) Identities() identities.Client {
	return t.ids
}

func (t *testBackends) Features() features.Client {
	return t.fc
}

type TestBackendOpts func(*testBackends)

func NewTestBackends(t *testing.T, opts ...TestBackendOpts) Backends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	tb := &testBackends{
		ac:   analytics_client.New(nil, ""),
		val:  validator.New(),
		keys: kc,
	}

	for _, opt := range opts {
		opt(tb)
	}

	return tb
}
