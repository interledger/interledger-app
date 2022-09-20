package graph

import (
	"context"
	"net/http/httptest"
	"testing"

	user_mock "gitlab.com/fynbos/backend/user/client/mock"

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
	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"
	country_client "gitlab.com/fynbos/backend/country/client"
	funding_client "gitlab.com/fynbos/backend/fundingsources/client"
	"gitlab.com/fynbos/backend/identity"
	identity_client "gitlab.com/fynbos/backend/identity/client"
	"gitlab.com/fynbos/backend/onboarding"
	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"
	mx_mock "gitlab.com/fynbos/backend/providers/mx/client/mock"
	"gitlab.com/fynbos/pacioli"
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
	CountryService       _country.Client
	FundingSourceService fundingsources.Client
	IdentityService      identity.Client
	UserService          _user.Client
	Mx                   mx.Client
	NoopService          _noop.Service
	UnitService          unit.Client
	UnitMockServer       *httptest.Server
	TransactionService   account_transactions.Client
	Os                   onboarding.Client
	PacioliClient        pacioli.Client
	PacioliLedgerID      uint32
	Graph                *handler.Server
	Client               *graphql.Client
	Server               *httptest.Server
	ValidatorImpl        *validator.Validate
	Tp                   client.Client
}

func (c *TestContainer) Noop() _noop.Service {
	return c.NoopService
}

func (c *TestContainer) Temporal() client.Client {
	return c.Tp
}

func (c *TestContainer) Unit() unit.Client {
	return c.UnitService
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

func (c *TestContainer) Identity() identity.Client {
	return c.IdentityService
}

func (c *TestContainer) Countries() _country.Client {
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

	cs := country_client.New(c)
	c.CountryService = cs

	is := identity_client.New(c, logger)
	c.IdentityService = is

	c.PacioliLedgerID = 1
	pClient := test_utils.SetupPacioli(t, ctx)

	c.PacioliClient = pClient

	as := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.AccountService = as

	users := user_mock.NewMock()
	c.UserService = users

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

	tp := &mocks.Client{}
	c.Tp = tp

	c.Mx = mx_mock.NewMockClient(c.Ctrl)

	fs := funding_client.New(c, logger)
	c.FundingSourceService = fs

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

	os := onboarding_client.New(c)
	c.Os = os

	graph, err := NewService(GraphqlOpts{
		Db:                  db,
		Identity:            is,
		User:                users,
		Account:             as,
		Country:             cs,
		Noop:                noopProvider,
		AccountTransactions: ts,
		Os:                  os,
		Fs:                  fs,
	})
	graph = NewLoggingService(graph, logger)
	c.Graph = graph

	router := chi.NewRouter()
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
