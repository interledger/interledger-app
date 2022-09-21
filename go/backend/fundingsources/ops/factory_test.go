package ops_test

import (
	"context"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	funding_client "gitlab.com/fynbos/backend/fundingsources/client"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

type TestContainer struct {
	Ctx           context.Context
	Logger        *zap.Logger
	Db            *sqlx.DB
	Us            user.Client
	Fs            fundingsources.Client
	ValidatorImpl *validator.Validate
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

	c.Us = user_client.New(c, "kratos")

	fs := funding_client.New(c, logger)
	c.Fs = fs

	return c, nil
}
