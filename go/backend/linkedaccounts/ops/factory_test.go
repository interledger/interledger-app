package ops_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"testing"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	linked_account_client "gitlab.com/fynbos/backend/linkedaccounts/client"
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
	Ac             analytics.Client
	kc             keys.Client
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Users() user.Client {
	return t.Us
}

func (t TestContainer) Notify() notify.Client {
	return t.Nf
}

func (t TestContainer) Analytics() analytics.Client {
	return t.Ac
}

func (t TestContainer) Keys() keys.Client {
	return t.kc
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

	c.Ac = analytics_client.New(c, "")

	c.Us = user_client.New(c, "kratosURL", "kratosAdminURL")

	c.LinkedAccounts = linked_account_client.New(c)

	c.Nf = notify_client.New(c, "")

	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	c.kc = kc

	return c, nil
}
