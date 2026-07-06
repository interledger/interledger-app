package ops

import (
	"testing"

	payments_mock "github.com/interledger/interledger-app/go/backend/payments/client/mock"

	"github.com/interledger/interledger-app/go/backend/payments"

	"github.com/interledger/interledger-app/go/backend/email"
	email_mock "github.com/interledger/interledger-app/go/backend/email/client/mock"
	"github.com/interledger/interledger-app/go/backend/kyc"
	kyc_mock "github.com/interledger/interledger-app/go/backend/kyc/client/mock"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"
	"github.com/interledger/interledger-app/go/backend/notify"
	notify_client "github.com/interledger/interledger-app/go/backend/notify/client/mock"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Users() user.Client
	Wallets() wallets.Client
	Notify() notify.Client
	Analytics() analytics.Client
	KYC() kyc.Client
	Email() email.Client
	Payments() payments.Client
}

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	notify notify.Client
	ac     analytics.Client
	ec     *email_mock.MockClient
	kyc    *kyc_mock.MockClient
	pc     *payments_mock.MockClient
	wc     *wallets_mock.MockClient
}

func (t testBackends) Payments() payments.Client {
	return t.pc
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) Wallets() wallets.Client {
	return t.wc
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

func (t testBackends) KYC() kyc.Client {
	return t.kyc
}

func (t testBackends) Email() email.Client {
	return t.ec
}

func NewTestBackends(t *testing.T, db *sqlx.DB) Backends {
	ctrl := gomock.NewController(t)
	nc := notify_client.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	ec := email_mock.NewMockClient(ctrl)
	ec.EXPECT().SendConnectedAccountEmail(gomock.Any(), gomock.Any()).AnyTimes()
	kycMock := kyc_mock.NewMockClient(ctrl)
	kycMock.EXPECT().GetIndividualDetails(gomock.Any(), gomock.Any()).Return(&kyc.IndividualDetails{}, nil).AnyTimes()
	pc := payments_mock.NewMockClient(ctrl)
	pc.EXPECT().SignalAccountLinked(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().GetFromAddress(gomock.Any(), gomock.Any()).AnyTimes()
	return &testBackends{db: db, val: validator.New(), notify: nc, ac: analytics_client.New(nil, ""), ec: ec, kyc: kycMock, pc: pc, wc: wc}
}
