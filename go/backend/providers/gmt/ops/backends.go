package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/aws"

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
	AWS() aws.Client
}

type TestBackends struct {
	Db              *sqlx.DB
	AnalyticsClient analytics.Client
	KeysClient      keys.Client
	AWSImpl         aws.Client
}

func (t TestBackends) AWS() aws.Client {
	return t.AWSImpl
}

func (t TestBackends) Limits() limits.Client {
	return nil
}

func (t TestBackends) MX() mx.Client {
	return nil
}

func (t TestBackends) Analytics() analytics.Client {
	return t.AnalyticsClient
}

func (t TestBackends) DB() *sqlx.DB {
	return t.Db
}

func (t TestBackends) Users() user.Client {
	return nil
}

func (t TestBackends) KYC() kyc.Client {
	return nil
}

func (t TestBackends) LinkedAccounts() linkedaccounts.Client {
	return nil
}

func (t TestBackends) Temporal() client.Client {
	return nil
}

func (t TestBackends) Transactions() transactions.Client {
	return nil
}

func (t TestBackends) Validator() *validator.Validate {
	return nil
}

func (t TestBackends) Keys() keys.Client {
	return t.KeysClient
}

func (t TestBackends) Tabapay() tabapay.Client {
	return nil
}

func NewTestBackends(t *testing.T, db *sqlx.DB, opts ...func(b *TestBackends)) Backends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	b := &TestBackends{Db: db, AnalyticsClient: analytics_client.New(nil, ""), KeysClient: kc}

	for _, opt := range opts {
		opt(b)
	}

	return b
}
