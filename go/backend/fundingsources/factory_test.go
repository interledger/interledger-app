package fundingsources

import (
	"context"
	"testing"

	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type TestContainer struct {
	Ctx              context.Context
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
	PacioliLedgerID  uint32
	Tp               *mocks.Client
	Unit             *_unit.MockService
}

func NewTestContainer(ctx context.Context, t *testing.T, ctrl *gomock.Controller) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(t, ctx)
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
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
	c.Is = _identity.NewLoggingService(is, logger)

	c.PacioliContainer = test_utils.SetupPacioli(t, ctx)

	c.PacioliLedgerID = 1
	conn, err := grpc.Dial(c.PacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		EquityAccID:   "46d4b2bd-e29b-4a63-9aa8-7990776c714e",
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}

	c.Noop = noop
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

	c.Unit = _unit.NewMockService(ctrl)
	fs, err := NewService(&ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: noop,
		Tp:   c.Tp,
		Unit: c.Unit,
	})
	if err != nil {
		return nil, err
	}
	c.Fs = NewLoggingService(fs, logger)

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
