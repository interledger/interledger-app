package graph

import (
	"context"
	"net/http/httptest"
	"testing"

	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/mock"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/withdrawals"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/accounts"
	_account "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/providers/mx"
	_noop "gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
)

type TestContainer struct {
	Ctx                  context.Context
	Ctrl                 *gomock.Controller
	Logger               *zap.Logger
	Db                   *sqlx.DB
	AccountService       _account.Client
	CountryService       _country.Service
	FundingSourceService fundingsources.Service
	IdentityService      identity.Service
	UserService          _user.Service
	Mx                   *mx.MockService
	NoopService          _noop.Service
	UnitService          unit.Service
	UnitMockServer       *httptest.Server
	DepositService       deposits.Service
	WithdrawalService    withdrawals.Service
	TransactionService   account_transactions.Client
	Os                   onboarding.Service
	Ps                   payments.Service
	PacioliContainer     *test_utils.PacioliContainer
	PacioliClient        pacioli.Client
	PacioliLedgerID      uint32
	Graph                *handler.Server
	Client               *graphql.Client
	Server               *httptest.Server
	ValidatorImpl        *validator.Validate
}

func (c *TestContainer) Accounts() _account.Client {
	return c.AccountService
}

func (c *TestContainer) Validator() *validator.Validate {
	return c.ValidatorImpl
}

func (c *TestContainer) DB() *sqlx.DB {
	return c.Db
}

func (c *TestContainer) Identity() identity.Service {
	return c.IdentityService
}

func (c *TestContainer) Countries() _country.Service {
	return c.CountryService
}

func (c *TestContainer) Pacioli() pacioli.Client {
	return c.PacioliClient
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	c.Ctrl = gomock.NewController(t)

	db := test_utils.MigrateCockroachDB(t, ctx)
	c.Db = db

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
		return nil, err
	}
	c.IdentityService = identity.NewLoggingService(is, logger)

	c.PacioliContainer = test_utils.SetupPacioli(t, ctx)

	c.PacioliLedgerID = 1
	pClient, err := pacioli_client.New(c.PacioliContainer.PacioliUrl)
	if err != nil {
		return nil, err
	}

	c.PacioliClient = pClient

	as := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.AccountService = as

	users := _user.NewMockService()
	c.UserService = _user.NewLoggingService(users, logger)

	ts := transactions_client.New(c, logger)
	c.TransactionService = ts

	equityAccID := "46d4b2bd-e29b-4a63-9aa8-7990776c714e"
	noopProvider, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   equityAccID,
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}

	c.NoopService = noopProvider

	c.UnitMockServer = test_utils.SetupUnitMockServer(ctx)

	us, err := unit.NewService(unit.ServiceArgs{
		BaseURL:         c.UnitMockServer.URL,
		WebhookToken:    "test-webhook-token",
		Token:           "test token",
		Db:              db,
		IdentityService: is,
		Logger:          logger,
	})
	if err != nil {
		return nil, err
	}
	c.UnitService = us

	tp := &mocks.Client{}
	c.Mx = mx.NewMockService(c.Ctrl)
	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: noopProvider,
		Unit: us,
		Tp:   tp,
	})
	if err != nil {
		return nil, err
	}
	c.FundingSourceService = fundingsources.NewLoggingService(fs, logger)

	tp.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, opts client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WorkflowRun {
			testWorkflowID := opts.ID
			testRunID := "test-runid"

			mockWorkflowRun := &mocks.WorkflowRun{}
			mockWorkflowRun.On("GetID").Return(testWorkflowID)
			mockWorkflowRun.On("GetRunID").Return(testRunID)
			mockWorkflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
			return mockWorkflowRun
		}, nil,
	)

	ds, err := deposits.NewService(&deposits.ServiceArgs{
		Db: db,
		As: as,
		Is: is,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		return nil, err
	}
	c.DepositService = ds

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: noopProvider,
		Tp:   tp,
	})
	if err != nil {
		return nil, err
	}
	c.Os = os

	ws, err := withdrawals.NewService(&withdrawals.ServiceArgs{
		Db: db,
		As: as,
		Is: is,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		return nil, err
	}
	c.WithdrawalService = ws

	ps, err := payments.NewService(&payments.ServiceArgs{
		Db: db,
		As: as,
		Is: is,
		Tp: tp,
	})
	if err != nil {
		return nil, err
	}
	c.Ps = payments.NewLoggingService(ps, logger)

	graph, err := NewService(GraphqlOpts{
		Db:                  db,
		Identity:            is,
		User:                users,
		Account:             as,
		Country:             cs,
		Noop:                noopProvider,
		UnitService:         us,
		AccountTransactions: ts,
		Ds:                  ds,
		Os:                  os,
		Ws:                  ws,
		Fs:                  fs,
		Ps:                  ps,
	})
	graph = NewLoggingService(graph, logger)
	c.Graph = graph

	router := chi.NewRouter()
	router.Use(_user.MakeMiddleware(users))
	router.Handle("/graphql", MakeHandler(graph, GraphqlHttpHandlerOpts{}))
	server := httptest.NewServer(router)
	c.Server = server

	c.Client = graphql.NewClient(server.URL + "/graphql")
	return c, err
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	c.Server.Close()

	c.UnitMockServer.Close()

	return nil
}

func NewIdentity(
	container *TestContainer,
	input *identity.CreateArgs,
) (*identity.Identity, error) {
	id, err := container.IdentityService.Create(container.Ctx, input)
	if err != nil {
		return nil, err
	}

	return id, nil
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

func NewBankAccount(
	container *TestContainer,
	user *_user.User,
	args *fundingsources.CreateBankAccountArgs,
	verify bool,
) (*fundingsources.FundingSource, error) {
	bankAccount, err := container.FundingSourceService.CreateBankAccount(container.Ctx, args)
	if err != nil {
		return nil, err
	}

	if verify {
		bankAccount, err = container.FundingSourceService.Verify(
			container.Ctx,
			&fundingsources.VerifyArgs{
				IdentityID:      user.ID,
				FundingSourceID: bankAccount.ID,
			})
		if err != nil {
			return nil, err
		}
	}

	return bankAccount, nil
}

func NewDeposit(
	c *TestContainer,
	args *account_transactions.CreateTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	trx, err := c.TransactionService.Create(c.Ctx, args)
	if err != nil {
		return nil, err
	}

	return trx, nil
}

func NewOutgoingPayment(
	c *TestContainer,
	args *account_transactions.CreateTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	trx, err := c.TransactionService.Create(c.Ctx, args)
	if err != nil {
		return nil, err
	}

	return trx, nil
}
