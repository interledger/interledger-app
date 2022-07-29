package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/fundingsources"

	country_client "gitlab.com/fynbos/backend/country/client"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"
	"gitlab.com/fynbos/backend/country"
	funding_client "gitlab.com/fynbos/backend/fundingsources/client"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

type TestContainer struct {
	Ctx              context.Context
	Logger           *zap.Logger
	Db               *sqlx.DB
	Cs               country.Client
	NoopImpl         noop.Service
	Is               _identity.Service
	Fs               fundingsources.Client
	Os               onboarding.Service
	Ts               account_transactions.Client
	as               accounts.Client
	PacioliContainer *test_utils.PacioliContainer
	PacioliClient    pacioli.Client
	PacioliLedgerID  uint32
	Tp               *mocks.Client
	UnitImpl         *_unit.MockService
	ValidatorImpl    *validator.Validate
}

func (t TestContainer) Noop() noop.Service {
	return t.NoopImpl
}

func (t TestContainer) Temporal() client.Client {
	return t.Tp
}

func (t TestContainer) Unit() _unit.Service {
	return t.UnitImpl
}

func (t TestContainer) Accounts() accounts.Client {
	return t.as
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Identity() _identity.Service {
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

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
	c.Is = _identity.NewLoggingService(is, logger)

	c.PacioliContainer = test_utils.SetupPacioli(t, ctx)

	c.PacioliLedgerID = 1

	pClient, err := pacioli_client.New(c.PacioliContainer.PacioliUrl)
	if err != nil {
		return nil, err
	}
	c.PacioliClient = pClient

	as := accounts_client.New(c, c.PacioliLedgerID, logger)
	c.as = as

	ts := transactions_client.New(c, logger)
	c.Ts = ts

	noop, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   "46d4b2bd-e29b-4a63-9aa8-7990776c714e",
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}

	c.NoopImpl = noop
	c.Tp = &mocks.Client{}
	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: noop,
		Tp:   c.Tp,
	})
	if err != nil {
		return nil, err
	}
	c.Os = os

	c.UnitImpl = _unit.NewMockService(ctrl)
	fs := funding_client.New(c, logger)
	c.Fs = fs

	return c, nil
}

func NewAccount(
	container *TestContainer,
	input *onboarding.CreateAccountArgs,
) (*accounts.Account, error) {
	acc, err := container.Os.CreateAccount(container.Ctx, input)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
