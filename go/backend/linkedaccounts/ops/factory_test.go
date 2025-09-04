package ops_test

import (
	"context"
	"testing"

	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"

	"gitlab.com/fynbos/backend/email"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_account_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"gitlab.com/fynbos/backend/user"
	"go.uber.org/zap"
)

type TestContainer struct {
	Ctx            context.Context
	Logger         *zap.Logger
	Db             *sqlx.DB
	Us             user.Client
	LinkedAccounts linkedaccounts.Client
	ValidatorImpl  *validator.Validate
	Nf             notify.Client
	kc             keys.Client
	Wc             *wallets_mock.MockClient
	Ec             *email_mock.MockClient
	Kyc            *kyc_mock.MockClient
	Pc             *payments_mock.MockClient
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Wallets() wallets.Client {
	return t.Wc
}

func (t TestContainer) Notify() notify.Client {
	return t.Nf
}

func (t TestContainer) Keys() keys.Client {
	return t.kc
}

func (t TestContainer) KYC() kyc.Client {
	return t.Kyc
}

func (t TestContainer) Email() email.Client {
	return t.Ec
}

func (t TestContainer) Payments() payments.Client {
	return t.Pc
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	mdb := db.MigrateTestDB(t, ctx)
	c.Db = mdb

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.LinkedAccounts = linked_account_client.New(c)

	c.Nf = notify_client.New(c, "")

	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	wc := wallets_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	c.kc = kc
	c.Wc = wc
	c.Ec = email_mock.NewMockClient(ctrl)
	c.Kyc = kyc_mock.NewMockClient(ctrl)
	c.Pc = payments_mock.NewMockClient(ctrl)

	c.Kyc.EXPECT().GetIndividualDetails(gomock.Any(), gomock.Any()).Return(&kyc.IndividualDetails{
		FirstName: "Test",
		LastName:  "User",
	}, nil).AnyTimes()
	c.Pc.EXPECT().SignalAccountLinked(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return c, nil
}
