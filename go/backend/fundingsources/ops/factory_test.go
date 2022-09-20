package ops_test

import (
	"context"
	"testing"

	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"

	identity_client "gitlab.com/fynbos/backend/identity/client"

	"gitlab.com/fynbos/backend/fundingsources"

	country_client "gitlab.com/fynbos/backend/country/client"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/country"
	funding_client "gitlab.com/fynbos/backend/fundingsources/client"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

type TestContainer struct {
	Ctx             context.Context
	Logger          *zap.Logger
	Db              *sqlx.DB
	Cs              country.Client
	Is              _identity.Client
	Fs              fundingsources.Client
	Os              onboarding.Client
	PacioliClient   pacioli.Client
	PacioliLedgerID uint32
	Tp              *mocks.Client
	ValidatorImpl   *validator.Validate
}

func (t TestContainer) Temporal() client.Client {
	return t.Tp
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Identity() _identity.Client {
	return t.Is
}

func (t TestContainer) Countries() country.Client {
	return t.Cs
}

func (t TestContainer) Pacioli() pacioli.Client {
	return t.PacioliClient
}

func NewTestContainer(ctx context.Context, t *testing.T, ctrl *gomock.Controller) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(t, ctx)
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := country_client.New(c)
	c.Cs = cs

	is := identity_client.New(c, logger)
	c.Is = is

	c.PacioliLedgerID = 1
	pClient := test_utils.SetupPacioli(t, ctx)
	c.PacioliClient = pClient

	c.Tp = &mocks.Client{}

	os := onboarding_client.New(c)
	c.Os = os

	fs := funding_client.New(c, logger)
	c.Fs = fs

	return c, nil
}
