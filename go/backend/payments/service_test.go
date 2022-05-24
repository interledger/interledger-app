package payments

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	_country "gitlab.com/fynbos/backend/country"
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

func TestPayments(s *testing.T) {
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

	s.Run("initiating an outgoing payment adds a workflow", func(t *testing.T) {
		amount := uint64(100)
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
		acc, err := NewVerifiedAccount(
			container,
			&onboarding.CreateAccountArgs{
				IdentityID: id.ID,
				Country:    id.Country,
			},
			&onboarding.VerifyAccountArgs{
				DateOfBirth: faker.Date(),
				Address:     []string{faker.Name()},
				State:       faker.Name(),
				City:        faker.Name(),
				PostalCode:  faker.CCNumber(),
				TaxIDNumber: faker.CCNumber(),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = NewDeposit(container, &account_transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Test transaction",
			Type:        "deposit",
			NetAmount:   1000,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        container.NoopService.GetLedgerID(),
					CreditAccountID: container.NoopService.GetEquityAccountID(),
					DebitAccountID:  acc.LedgerAccountID,
					Amount:          1000,
					// Code: uint16,
					Flags: account_transactions.LedgerTransferFlags{},
				},
			},
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

		p, err := container.PaymentService.InitiateOutgoingPayment(context.Background(), &InitiateOutgoingPaymentArgs{
			IdentityID: acc.IdentityID,
			AccountID:  acc.ID,
			Amount:     amount,
			To:         "$test.fynbos.test/alice",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, Created, p.State)
		assert.Equal(t, acc.ID, p.AccountID)
		assert.Equal(t, amount, p.Amount)
	})

	s.Run("initiating an outgoing payment with insufficient balance fails", func(t *testing.T) {
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
		acc, err := NewVerifiedAccount(
			container,
			&onboarding.CreateAccountArgs{
				IdentityID: id.ID,
				Country:    id.Country,
			},
			&onboarding.VerifyAccountArgs{
				DateOfBirth: faker.Date(),
				Address:     []string{faker.Name()},
				State:       faker.Name(),
				City:        faker.Name(),
				PostalCode:  faker.CCNumber(),
				TaxIDNumber: faker.CCNumber(),
			},
		)
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

		p, err := container.PaymentService.InitiateOutgoingPayment(context.Background(), &InitiateOutgoingPaymentArgs{
			IdentityID: acc.IdentityID,
			AccountID:  acc.ID,
			Amount:     100,
			To:         "$test.fynbos.test/alice",
		})

		assert.Nil(t, p)
		assert.ErrorIs(t, err, ErrInsufficientBalance)
	})
}

type TestContainer struct {
	IdentityService    _identity.Service
	AccountService     _accounts.Service
	CountryService     _country.Service
	NoopService        noop.Service
	OnboardService     onboarding.Service
	TransactionService account_transactions.Service
	PaymentService     Service
	TemporalMock       *mocks.Client
	PacioliContainer   *test_utils.PacioliContainer
	PacioliClient      pacioliv1.PacioliServiceClient
	PacioliLedgerID    uint16
	Db                 *sqlx.DB
	DbCleanup          func()
	Logger             *zap.Logger
	Ctx                context.Context
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
	db, cleanup := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db
	c.DbCleanup = cleanup

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

	temporal := &mocks.Client{}
	c.TemporalMock = temporal

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: np,
		Tp:   temporal,
	})
	if err != nil {
		return nil, err
	}
	c.OnboardService = os

	ts, err := account_transactions.NewService(&account_transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}

	c.TransactionService = ts

	ps, err := NewService(&ServiceArgs{
		Db: db,
		Is: is,
		As: as,
		Tp: temporal,
	})
	if err != nil {
		return nil, err
	}

	c.PaymentService = ps

	return c, nil
}

// Funcion will set the `AccountID` and `IdentityID` on the `verifyAccountArgs` for you.
func NewVerifiedAccount(
	container *TestContainer,
	createAccountArgs *onboarding.CreateAccountArgs,
	verifyAccountArgs *onboarding.VerifyAccountArgs,
) (*_accounts.Account, error) {
	acc, err := container.OnboardService.CreateAccount(container.Ctx, createAccountArgs)
	if err != nil {
		return nil, err
	}

	verifyAccountArgs.AccountID = acc.ID
	verifyAccountArgs.IdentityID = acc.IdentityID
	acc, err = container.OnboardService.VerifyAccount(container.Ctx, verifyAccountArgs)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func NewDeposit(
	container *TestContainer,
	args *account_transactions.CreateTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	trx, err := container.TransactionService.Create(container.Ctx, args)
	if err != nil {
		return nil, err
	}

	return trx, nil
}
