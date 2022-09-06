package withdrawals

import (
	"context"
	"errors"
	"fmt"
	"testing"

	mx_mock "gitlab.com/fynbos/backend/providers/mx/client/mock"

	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"

	identity_client "gitlab.com/fynbos/backend/identity/client"

	funding_client "gitlab.com/fynbos/backend/fundingsources/client"

	country_client "gitlab.com/fynbos/backend/country/client"

	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"

	"github.com/bxcodec/faker/v3"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_mock "gitlab.com/fynbos/backend/providers/unit/client/mock"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
)

func TestWithdrawals(s *testing.T) {
	ctx := context.Background()
	container, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("initiating a withdrawal adds a workflow", func(t *testing.T) {
		user := &_user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		amount := uint64(100)
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
		acc, err := NewAccount(container, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = container.TransactionService.Create(ctx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Random Deposit",
			Type:        "deposit",
			NetAmount:   1000,
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.PacioliLedgerID,
					DebitAccountID:  acc.LedgerAccountID,
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					Amount:          1000,
					Code:            1,
					Flags:           transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		fs, err := container.FundingSourcesService.CreateBankAccount(ctx, &fundingsources.CreateBankAccountArgs{
			IdentityID:    user.ID,
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
			IdentityID:      user.ID,
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

		w, err := container.WithdrawalService.InitiateWithdrawal(context.Background(), &InitiateWithdrawalArgs{
			IdentityID:      user.ID,
			AccountID:       acc.ID,
			FundingSourceID: fs.ID,
			Amount:          amount,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, Created, w.State)
		assert.Equal(t, acc.ID, w.AccountID)
		assert.Equal(t, fs.ID, w.FundingSourceId)
		assert.Equal(t, amount, w.Amount)
	})

	s.Run("insufficient balance wont create a withdrawal workflow", func(t *testing.T) {
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
			IdentityID:    user.ID,
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
			IdentityID:      user.ID,
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

		_, err = container.WithdrawalService.InitiateWithdrawal(context.Background(), &InitiateWithdrawalArgs{
			IdentityID:      user.ID,
			AccountID:       acc.ID,
			FundingSourceID: fs.ID,
			Amount:          amount,
		})
		if err == nil {
			t.Fatal(fmt.Errorf("error should occured"))
		}

		assert.True(t, errors.Is(err, ErrInsufficientBalance))
	})
}

type TestContainer struct {
	Ctrl                  *gomock.Controller
	IdentityService       identity.Client
	AccountService        _accounts.Client
	TransactionService    transactions.Client
	CountryService        _country.Client
	NoopService           noop.Service
	OnboardService        onboarding.Client
	FundingSourcesService fundingsources.Client
	WithdrawalService     Service
	Mx                    *mx_mock.MockClient
	UnitImpl              *unit_mock.MockClient
	TemporalMock          *mocks.Client
	PacioliClient         pacioli.Client
	PacioliLedgerID       uint32
	Db                    *sqlx.DB
	Logger                *zap.Logger
	Ctx                   context.Context
	ValidatorImpl         *validator.Validate
}

func (t TestContainer) Noop() noop.Service {
	return t.NoopService
}

func (t TestContainer) Temporal() client.Client {
	return t.TemporalMock
}

func (t TestContainer) Unit() unit.Client {
	return t.UnitImpl
}

func (t TestContainer) Accounts() _accounts.Client {
	return t.AccountService
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Identity() identity.Client {
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

	c.PacioliLedgerID = 1

	pClient := test_utils.SetupPacioli(s, ctx)
	c.PacioliClient = pClient

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := country_client.New(c)
	c.CountryService = cs

	is := identity_client.New(c, logger)
	c.IdentityService = is

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
	temporal := &mocks.Client{}
	c.TemporalMock = temporal
	c.UnitImpl = unit_mock.NewMockClient(c.Ctrl)
	c.Mx = mx_mock.NewMockClient(c.Ctrl)

	fs := funding_client.New(c, logger)
	c.FundingSourcesService = fs

	os := onboarding_client.New(c)
	c.OnboardService = os

	ws, err := NewService(&ServiceArgs{
		Db: db,
		Is: is,
		As: as,
		Fs: fs,
		Tp: temporal,
	})
	if err != nil {
		return nil, err
	}
	c.WithdrawalService = ws

	at := transactions_client.New(c, logger)
	c.TransactionService = at

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
