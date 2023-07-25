package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/images"
	images_mock "gitlab.com/fynbos/backend/images/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	openpayments_mock "gitlab.com/fynbos/backend/openpayments/client/mock"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/twitter"
	twitter_mock "gitlab.com/fynbos/backend/twitter/client/mock"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() temporal.Client
	Analytics() analytics.Client
	Keys() keys.Client
	OpenPayments() openpayments.Client
	Images() images.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	an  analytics.Client
	kc  *keys_mock.MockClient
	tc  *twitter_mock.MockClient
	op  openpayments.Client
	img images.Client
}

func (t testBackends) Twilio() twilio.Service {
	return nil
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	panic("implement me")
}

func (t testBackends) Transactions() transactions.Client {
	panic("implement me")
}

func (t testBackends) Tabapay() tabapay.Client {
	panic("implement me")
}

func (t testBackends) Temporal() temporal.Client {
	panic("implement me")
}

func (t testBackends) Twitter() twitter.Client {
	return t.tc
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Analytics() analytics.Client {
	return t.an
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) OpenPayments() openpayments.Client {
	return t.op
}

func (t testBackends) Images() images.Client {
	return t.img
}

func (t testBackends) Features() features.Client {
	return nil
}

func NewTestBackends(t *testing.T, db *sqlx.DB) *testBackends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	tc := twitter_mock.NewMockClient(ctrl)
	img := images_mock.NewMockClient(ctrl)
	img.EXPECT().GenerateTwitterIdentity(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	img.EXPECT().GenerateTwitterIdentityOG(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	kc.EXPECT().List(gomock.Any(), gomock.Any()).Return([]keys.Key{
		{
			ID:   "test",
			Type: keys.Custodial,
		},
	}, nil).AnyTimes()
	kc.EXPECT().Sign(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	op := openpayments_mock.NewMockClient(ctrl)
	op.EXPECT().GetWalletPaymentPointer(gomock.Any(), gomock.Any()).Return(&openpayments.PaymentPointer{
		ID:         "",
		URL:        "",
		WalletID:   "",
		Alias:      "",
		Asset:      "",
		AssetScale: 0,
	}, nil).AnyTimes()
	return &testBackends{db: db, val: validator.New(), an: analytics_client.New(nil, ""), kc: kc, tc: tc, op: op, img: img}
}
