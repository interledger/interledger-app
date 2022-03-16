package fundingsources

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type TestContainer struct {
	Ctx              context.Context
	Crdb             *test_utils.CockroachDBContainer
	Logger           *zap.Logger
	Db               *sqlx.DB
	Cs               country.Service
	Noop             noop.Service
	Is               _identity.Service
	Fs               Service
	Os               onboarding.Service
	Ts               account_transactions.Service
	PacioliContainer *test_utils.PacioliContainer
	PacioliClient    pacioliv1.PacioliServiceClient
	PacioliLedgerID  uint16
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := country.NewService(db)
	c.Cs = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
	})
	if err != nil {
		return nil, err
	}
	c.Is = _identity.NewLoggingService(is, logger)

	pacioliContainer, err := test_utils.SetupPacioli(ctx)
	if err != nil {
		return nil, err
	}
	c.PacioliContainer = pacioliContainer

	c.PacioliLedgerID = uint16(1)
	conn, err := grpc.Dial(pacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	pClient := pacioliv1.NewPacioliServiceClient(conn)
	c.PacioliClient = pClient
	as, err := accounts.NewService(&accounts.ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: c.PacioliLedgerID,
		PacioliClient:   pClient,
		Db:              db,
	})
	if err != nil {
		return nil, err
	}
	err = as.Init(ctx)
	if err != nil {
		return nil, err
	}

	ts, err := account_transactions.NewService(&account_transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
	c.Ts = ts

	noop, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   uuid.NewString(),
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}
	err = noop.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.Noop = noop

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: noop,
	})
	if err != nil {
		return nil, err
	}
	c.Os = os

	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: noop,
	})
	if err != nil {
		return nil, err
	}
	c.Fs = NewLoggingService(fs, logger)

	return c, nil
}

func (c *TestContainer) Cleanup() error {
	err := c.Db.Close()
	if err != nil {
		return err
	}

	err = c.Crdb.Container.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	err = c.PacioliContainer.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
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
