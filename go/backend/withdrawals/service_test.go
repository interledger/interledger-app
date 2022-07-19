package withdrawals

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitlab.com/fynbos/backend/accounts/ops"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_accounts "gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	IdentityService       identity.Service
	AccountService        ops.Service
	TransactionService    transactions.Service
	CountryService        _country.Service
	NoopService           noop.Service
	OnboardService        onboarding.Service
	FundingSourcesService fundingsources.Service
	WithdrawalService     Service
	Unit                  *unit.MockService
	Mx                    *mx.MockService
	TemporalMock          *mocks.Client
	PacioliContainer      *test_utils.PacioliContainer
	PacioliClient         pacioliv1.PacioliServiceClient
	PacioliLedgerID       uint32
	Db                    *sqlx.DB
	Logger                *zap.Logger
	Ctx                   context.Context
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	c.Ctrl = gomock.NewController(s)
	db := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db

	c.PacioliContainer = test_utils.SetupPacioli(s, ctx)

	c.PacioliLedgerID = 1
	conn, err := grpc.Dial(c.PacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	is, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = identity.NewLoggingService(is, logger)

	as, err := ops.NewService(&ops.ServiceArgs{
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
	c.AccountService = ops.NewLoggingService(as, logger)

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
	c.Unit = unit.NewMockService(c.Ctrl)
	c.Mx = mx.NewMockService(c.Ctrl)
	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Db:   db,
		Is:   is,
		As:   as,
		Noop: c.NoopService,
		Unit: c.Unit,
		Tp:   temporal,
	})
	if err != nil {
		return nil, err
	}
	c.FundingSourcesService = fs

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: np,
		Tp:   c.TemporalMock,
	})
	if err != nil {
		return nil, err
	}
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

	at, err := transactions.NewService(&transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
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
