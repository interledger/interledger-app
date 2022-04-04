package deposits

import (
	"context"
	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/pacioli/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	s.Run("initiating a deposit adds a workflow", func(t *testing.T) {
		userID := uuid.NewString()
		amount := uint64(100)
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID:   userID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        faker.Email(),
			Country:      "US",
		})
		container.MockPacioliClient.EXPECT().GetAccounts(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountsRequest, opts ...grpc.CallOption) (*pacioli.GetAccountsResponse, error) {
				return &pacioli.GetAccountsResponse{
					Accounts: []*pacioli.Account{
						{
							Id: args.Ids[0],
						},
					},
				}, nil
			}).AnyTimes()
		if err != nil {
			t.Fatal(err)
		}
		fs, err := container.FundingSourcesService.CreateBankAccount(ctx, &fundingsources.CreateBankAccountArgs{
			IdentityID:    userID,
			AccountID:     acc.ID,
			Name:          "my account",
			AccountNumber: "12345678",
			RoutingNumber: "1234",
			Institution:   "Bank",
			Type:          "cheque",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = container.FundingSourcesService.Verify(ctx, &fundingsources.VerifyArgs{
			IdentityID:      userID,
			FundingSourceID: fs.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		container.TemporalMock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
				testWorkflowID := opts.ID
				testRunID := "test-runid"

				mockWorkflowRun := &mocks.WorkflowRun{}
				mockWorkflowRun.On("GetID").Return(testWorkflowID)
				mockWorkflowRun.On("GetRunID").Return(testRunID)
				mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
				return mockWorkflowRun
			}, nil,
		).Times(1)

		d, err := container.DepositService.InitiateDeposit(context.Background(), &InitiateDepositArgs{
			IdentityID:      userID,
			AccountID:       acc.ID,
			FundingSourceID: fs.ID,
			Amount:          amount,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, Created, d.State)
		assert.Equal(t, acc.ID, d.AccountID)
		assert.Equal(t, fs.ID, d.FundingSourceId)
		assert.Equal(t, amount, d.Amount)
	})
}

type TestContainer struct {
	IdentityService       _identity.Service
	AccountService        _accounts.Service
	CountryService        _country.Service
	NoopService           noop.Service
	OnboardService        onboarding.Service
	FundingSourcesService fundingsources.Service
	DepositService        Service
	MockPacioliClient     *mockPacioliV1.MockPacioliServiceClient
	TemporalMock          *mocks.Client
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
	}).Return(&pacioli.ConfigureLedgersResponse{}, nil).Times(2)
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

	noop, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      pacioliLedgerID,
		EquityAccID:   uuid.NewString(),
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}
	c.MockPacioliClient.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
		&pacioli.ConfigureAccountsResponse{}, nil,
	).Times(1)
	err = noop.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.NoopService = noop

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Db:   db,
		Is:   is,
		As:   as,
		Noop: c.NoopService,
	})
	if err != nil {
		return nil, err
	}
	c.FundingSourcesService = fs

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: noop,
	})
	if err != nil {
		return nil, err
	}
	c.OnboardService = os

	temporal := &mocks.Client{}
	c.TemporalMock = temporal

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

func NewAccount(
	container *TestContainer,
	input *onboarding.CreateAccountArgs,
) (*_accounts.Account, error) {
	container.MockPacioliClient.EXPECT().ConfigureAccounts(gomock.Any(), gomock.Any()).Return(
		&pacioli.ConfigureAccountsResponse{}, nil,
	).Times(1)

	acc, err := container.OnboardService.CreateAccount(container.Ctx, input)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
