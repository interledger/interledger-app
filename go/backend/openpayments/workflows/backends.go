package workflows

import (
	"testing"

	"gitlab.com/fynbos/backend/twilio"

	features_mock "gitlab.com/fynbos/backend/features/client/mock"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"

	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/linkedaccounts"

	temporal "go.temporal.io/sdk/client"

	"gitlab.com/fynbos/backend/user"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	Users() user.Client
	KYC() kyc.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	Transactions() transactions.Client
	Analytics() analytics.Client
	Contacts() contacts.Client
	Tabapay() tabapay.Client
	Keys() keys.Client
	Features() features.Client
	Twilio() twilio.Service
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	t   temporal.Client
	la  linkedaccounts.Client
	em  email.Client
	tx  transactions.Client
	kyc kyc.Client
	ac  analytics.Client
	cc  contacts.Client
	tbc tabapay.Client
	kc  keys.Client
	fc  features.Client
	ts  twilio.Service
}

func (t testBackends) Twilio() twilio.Service {
	return t.ts
}

func (t testBackends) Transactions() transactions.Client {
	return t.tx
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Email() email.Client {
	return t.em
}

func (t testBackends) Temporal() temporal.Client {
	return t.t
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) Contacts() contacts.Client {
	return t.cc
}

func (t testBackends) Tabapay() tabapay.Client {
	return t.tbc
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) Features() features.Client {
	return t.fc
}

func NewTestBackends(t *testing.T, db *sqlx.DB, temp temporal.Client, la linkedaccounts.Client, tx transactions.Client, kyc kyc.Client, cc contacts.Client) Backends {
	ac := analytics_client.New(nil, "")
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	fc := features_mock.NewMockClient(ctrl)
	fc.EXPECT().Features(gomock.Any(), gomock.Any()).Return(&features.WalletFeatures{
		SendEnabled:       true,
		ReceiveEnabled:    true,
		LinkedAccEnabled:  true,
		CardsEnabled:      true,
		BanksEnabled:      true,
		IdentitiesEnabled: true,
		TwitterEnabled:    true,
	}, nil).AnyTimes()

	return &testBackends{db: db, val: validator.New(), t: temp, la: la, tx: tx, kyc: kyc, ac: ac, cc: cc, kc: kc, fc: fc}
}
