package deposits

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestDeposits(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err := container.Cleanup(ctx)
		if err != nil {
			return
		}
	})

	s.Run("initiating a deposit adds a workflow", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		id, err := container.IdentityService.Create(container.Ctx, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		amount := uint64(100)
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		fs, err := container.FundingSourcesService.CreateBankAccount(ctx, &fundingsources.CreateBankAccountArgs{
			IdentityID:    id.ID,
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
			IdentityID:      id.ID,
			FundingSourceID: fs.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		container.TemporalMock.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(
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
			IdentityID:      id.ID,
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
	TemporalMock          *mocks.Client
	PacioliContainer      *test_utils.PacioliContainer
	PacioliClient         pacioliv1.PacioliServiceClient
	PacioliLedgerID       uint16
	Db                    *sqlx.DB
	DbCleanup             func()
	Tp                    *mocks.Client
	Logger                *zap.Logger
	Ctx                   context.Context
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	c.DbCleanup()

	err := c.PacioliContainer.Terminate(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	db, dbCleanup := test_utils.MigrateCockroachDB(s, ctx)
	c.DbCleanup = dbCleanup
	c.Db = db

	pacioliContainer, err := test_utils.SetupPacioli(ctx)
	if err != nil {
		return nil, err
	}
	c.PacioliContainer = pacioliContainer

	c.PacioliLedgerID = uint16(1)
	conn, err := grpc.Dial(pacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	pClient := pacioliv1.NewPacioliServiceClient(conn)
	c.PacioliClient = pClient

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

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

	as, err := _accounts.NewService(&_accounts.ServiceArgs{
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
	c.AccountService = _accounts.NewLoggingService(as, logger)

	np, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   uuid.NewString(),
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}
	err = np.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.NoopService = np

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
	c.Tp = &mocks.Client{}
	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: np,
		Tp:   c.Tp,
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

	acc, err := container.OnboardService.CreateAccount(container.Ctx, input)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
