package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	linked_account_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

type TestContainer struct {
	Ctx            context.Context
	Logger         *zap.Logger
	Db             *sqlx.DB
	Us             user.Client
	LinkedAccounts linkedaccounts.Client
	ValidatorImpl  *validator.Validate
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

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(t, ctx)
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.Us = user_client.New(c, "kratosURL", "kratosAdminURL")

	c.LinkedAccounts = linked_account_client.New(c, logger)

	return c, nil
}
