package ops_test

import (
	"testing"

	"github.com/interledger/interledger-app/go/backend/payments"

	"github.com/google/uuid"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/jmoiron/sqlx"

	"github.com/interledger/interledger-app/go/backend/keys"
	keys_mock "github.com/interledger/interledger-app/go/backend/keys/client/mock"
	"github.com/interledger/interledger-app/go/backend/kyc"
	kyc_mock "github.com/interledger/interledger-app/go/backend/kyc/client/mock"
	"github.com/interledger/interledger-app/go/backend/notify"
	notify_client "github.com/interledger/interledger-app/go/backend/notify/client/mock"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	ac     analytics.Client
	notify notify.Client
	user   user.Client
	kc     keys.Client
	kyc    *kyc_mock.MockClient
	wc     wallets.Client
}

func (t testBackends) Payments() payments.Client {
	return nil
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

func (t testBackends) Wallets() wallets.Client {
	return t.wc
}

func (t testBackends) Email() email.Client {
	return nil
}

func NewTestBackends(t *testing.T, db *sqlx.DB, uc user.Client) *testBackends {

	ctrl := gomock.NewController(t)
	nc := notify_client.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()

	kyc := kyc_mock.NewMockClient(ctrl)

	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().GetFromAddress(gomock.Any(), gomock.Any()).Return(&wallets.Wallet{ID: uuid.NewString()}, nil).AnyTimes()

	return &testBackends{db: db, val: validator.New(), notify: nc, ac: analytics_client.New(nil, ""), user: uc, kc: kc, kyc: kyc, wc: wc}
}
