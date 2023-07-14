package ops_test

import (
	"testing"

	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"

	openpayments_mock "gitlab.com/fynbos/backend/openpayments/client/mock"

	"gitlab.com/fynbos/backend/user"

	analytics_client "gitlab.com/fynbos/backend/analytics/client"

	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/notify"

	notify_client "gitlab.com/fynbos/backend/notify/client/mock"

	"gitlab.com/fynbos/backend/openpayments"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
)

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	ac     analytics.Client
	notify notify.Client
	user   user.Client
	op     openpayments.Client
	kc     keys.Client
	kyc    *kyc_mock.MockClient
}

func (t testBackends) OpenPayments() openpayments.Client {
	return t.op
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Notify() notify.Client {
	return t.notify
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func (t testBackends) Users() user.Client {
	return t.user
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

func NewTestBackends(t *testing.T, db *sqlx.DB, uc user.Client) *testBackends {
	ctrl := gomock.NewController(t)
	nc := notify_client.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	op := openpayments_mock.NewMockClient(ctrl)
	op.EXPECT().GetWalletPaymentPointer(gomock.Any(), gomock.Any()).Return(&openpayments.PaymentPointer{Asset: "USD", URL: "https://fynbos.me/bobby"}, nil).AnyTimes()
	op.EXPECT().GetPaymentPointer(gomock.Any(), gomock.Any()).Return(&openpayments.PaymentPointer{Asset: "USD", URL: "https://fynbos.me/bobby"}, nil).AnyTimes()

	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()

	kyc := kyc_mock.NewMockClient(ctrl)

	return &testBackends{db: db, val: validator.New(), notify: nc, ac: analytics_client.New(nil, ""), user: uc, op: op, kc: kc, kyc: kyc}
}
