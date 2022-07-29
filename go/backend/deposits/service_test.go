package deposits

import (
	"context"
	"testing"

	funding_client "gitlab.com/fynbos/backend/fundingsources/client"

	country_client "gitlab.com/fynbos/backend/country/client"

	"github.com/bxcodec/faker/v3"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

func TestDeposits(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

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
	Ctrl                  *gomock.Controller
	IdentityService       _identity.Service
	AccountService        _accounts.Client
	CountryService        _country.Client
	NoopService           noop.Service
	OnboardService        onboarding.Service
	FundingSourcesService fundingsources.Client
	DepositService        Service
	Mx                    *_mx.MockService
	UnitImpl              *_unit.MockService
	TemporalMock          *mocks.Client
	PacioliContainer      *test_utils.PacioliContainer
	PacioliClient         pacioli.Client
	PacioliLedgerID       uint32
	Db                    *sqlx.DB
	Tp                    *mocks.Client
	Logger                *zap.Logger
	Ctx                   context.Context
	ValidatorImpl         *validator.Validate
}

func (t TestContainer) Accounts() _accounts.Client {
	return t.AccountService
}

func (t TestContainer) Noop() noop.Service {
	return t.NoopService
}

func (t TestContainer) Temporal() client.Client {
	return t.TemporalMock
}

func (t TestContainer) Unit() _unit.Service {
	return t.UnitImpl
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Identity() _identity.Service {
	return t.IdentityService
}

func (t TestContainer) Countries() _country.Client {
	return t.CountryService
}

func (t TestContainer) Pacioli() pacioli.Client {
	return t.PacioliClient
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	c.Ctrl = gomock.NewController(s)
	db := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db

	c.PacioliContainer = test_utils.SetupPacioli(s, ctx)

	c.PacioliLedgerID = uint32(1)
	pClient, err := pacioli_client.New(c.PacioliContainer.PacioliUrl)
	if err != nil {
		return nil, err
	}
	c.PacioliClient = pClient

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := country_client.New(c)
	c.CountryService = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = _identity.NewLoggingService(is, logger)

	as := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.AccountService = as

	np, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   "46d4b2bd-e29b-4a63-9aa8-7990776c714e",
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}

	c.NoopService = np
	c.Mx = _mx.NewMockService(c.Ctrl)
	c.UnitImpl = _unit.NewMockService(c.Ctrl)
	c.Tp = &mocks.Client{}

	fs := funding_client.New(c, logger)
	c.FundingSourcesService = fs

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
