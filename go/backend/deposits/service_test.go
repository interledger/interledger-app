package deposits

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	_accounts "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	_identity "gitlab.com/fynbos/backend/identity"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"testing"
)

func TestDeposits(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err := container.Cleanup()
		if err != nil {
			return
		}
	})
}

type TestContainer struct {
	IdentityService       _identity.Service
	AccountService        _accounts.Service
	CountryService        _country.Service
	FundingSourcesService fundingsources.Service
	DepositService        Service
	MockPacioliClient     *mockPacioliV1.MockPacioliServiceClient
	Ctrl                  *gomock.Controller
	Db                    *sqlx.DB
	Logger                *zap.Logger
	Crdb                  *test_utils.CockroachDBContainer
	Ctx                   context.Context
}

func (c *TestContainer) Cleanup() error {
	err := c.Db.Close()
	if err != nil {
		return err
	}
	c.Ctrl.Finish()

	return nil
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	// the tests are run in serial. We use a global connection for
	// each of the tests.
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

	ctrl := gomock.NewController(s)
	c.Ctrl = ctrl

	cs := _country.NewService(db)
	c.CountryService = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = _identity.NewLoggingService(is, logger)

	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
	c.MockPacioliClient = pClient
	pacioliLedgerID := uint16(1)
	pClient.EXPECT().ConfigureLedgers(ctx, &pacioli.ConfigureLedgersRequest{
		Args: []*pacioli.Ledger{
			{
				Id:    uint32(pacioliLedgerID),
				Name:  "Fynbos ledger",
				Asset: "840", // US dollars
				Scale: 2,
			},
		},
	}).Return(&pacioli.ConfigureLedgersResponse{}, nil).Times(1)
	as, err := _accounts.NewService(&_accounts.ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: pacioliLedgerID,
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
	c.AccountService = _accounts.NewLoggingService(as, logger)

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Db:   db,
		Is:   is,
		As:   as,
		Noop: nil,
	})
	if err != nil {
		return nil, err
	}

	temporal := &mocks.Client{}

	ds, err := NewService(&ServiceArgs{
		Db: db,
		Is: is,
		As: as,
		Fs: fs,
		Tp: temporal,
	})
	if err != nil {
		return nil, err
	}
	c.DepositService = ds

	return c, nil
}
