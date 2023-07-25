package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"

	"github.com/golang/mock/gomock"

	"gitlab.com/fynbos/backend/features"

	"gitlab.com/fynbos/backend/analytics"

	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	features_mock "gitlab.com/fynbos/backend/features/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
	LinkedAccounts() linkedaccounts.Client
	Transactions() transactions.Client
	Tabapay() tabapay.Client
	Analytics() analytics.Client
	Features() features.Client
	Twilio() twilio.Service
	Users() user.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	la  linkedaccounts.Client
	tc  transactions.Client
	tbc tabapay.Client
	ac  analytics.Client
	fc  features.Client
}

func (t testBackends) Twilio() twilio.Service {
	//TODO implement me
	panic("implement me")
}

func (t testBackends) Users() user.Client {
	//TODO implement me
	panic("implement me")
}

func (t testBackends) Transactions() transactions.Client {
	return t.tc
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	return t.la
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Tabapay() tabapay.Client {
	return t.tbc
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) Features() features.Client {
	return t.fc
}

func NewTestBackends(t *testing.T, db *sqlx.DB, la linkedaccounts.Client, tc transactions.Client) Backends {
	ac := analytics_client.New(nil, "")
	ctrl := gomock.NewController(t)
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

	return &testBackends{db: db, val: validator.New(), la: la, tc: tc, ac: ac, fc: fc}
}
