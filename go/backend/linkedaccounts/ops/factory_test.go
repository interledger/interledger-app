package ops_test

import (
	"context"
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

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	db := db.MigrateTestDB(t, ctx)
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.Us = user_client.New(c, "kratosURL", "kratosAdminURL")

	c.LinkedAccounts = linked_account_client.New(c)

	c.Nf = notify_client.New(c, "")

	return c, nil
}
