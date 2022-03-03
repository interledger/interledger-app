package graph

import (
	"context"
	"net/http/httptest"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/withdrawals"

	"google.golang.org/grpc"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/accounts"
	_account "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	_noop "gitlab.com/fynbos/backend/providers/noop"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	mockPacioliV1 "gitlab.com/fynbos/proto/pacioli/v1/mock"
)

type TestContainer struct {
	Ctx                  context.Context
	Crdb                 *test_utils.CockroachDBContainer
	Logger               *zap.Logger
	Db                   *sqlx.DB
	Ctrl                 *gomock.Controller
	AccountService       _account.Service
	CountryService       _country.Service
	FundingSourceService fundingsources.Service
	IdentityService      identity.Service
	UserService          _user.Service
	NoopService          _noop.Service
	DepositService       deposits.Service
	WithdrawalService    withdrawals.Service
	TransactionService   account_transactions.Service
	Os                   onboarding.Service
	Ps                   payments.Service
	MockPacioliClient    *mockPacioliV1.MockPacioliServiceClient
	Graph                *handler.Server
	Client               *graphql.Client
	Server               *httptest.Server
}

func NewTestContainer(ctx context.Context, t gomock.TestReporter) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	db, err := sqlx.Connect("postgres", crdb.URI)
	c.Db = db

	ctrl := gomock.NewController(t)
	c.Ctrl = ctrl

	cs := _country.NewService()
	c.CountryService = cs

	is, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
	})
	if err != nil {
		return nil, err
	}
	c.IdentityService = identity.NewLoggingService(is, logger)

	pacioliLedgerID := uuid.NewString()
	ledgerCode := uint16(1)
	pClient := mockPacioliV1.NewMockPacioliServiceClient(ctrl)
	pClient.EXPECT().GetLedgerByCode(gomock.Any(), gomock.Any()).Return(&pacioliv1.Ledger{
		Id:   pacioliLedgerID,
		Code: uint32(ledgerCode),
	}, nil).Times(1)
	c.MockPacioliClient = pClient

	as, err := accounts.NewService(&accounts.ServiceArgs{
		Is:                is,
		Cs:                cs,
		PacioliLedgerCode: ledgerCode,
		PacioliClient:     pClient,
		Db:                db,
	})
	if err != nil {
		return nil, err
	}
	c.AccountService = accounts.NewLoggingService(as, logger)

	users := _user.NewMockService()
	c.UserService = _user.NewLoggingService(users, logger)

	ts, err := transactions.NewService(&transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
	})
	if err != nil {
		return nil, err
	}
	ts = account_transactions.NewLoggingService(ts, logger)
	c.TransactionService = ts

	ledgerID := uuid.NewString()
	equityAccID := uuid.NewString()
	noopProvider, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:    ledgerID,
		EquityAccID: equityAccID,
	})
	if err != nil {
		return nil, err
	}
	c.NoopService = noopProvider

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Is:   is,
		As:   as,
		Db:   db,
		Noop: noopProvider,
	})
	if err != nil {
		return nil, err
	}
	c.FundingSourceService = fundingsources.NewLoggingService(fs, logger)

	ds, err := deposits.NewService(&deposits.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Fs:   fs,
		Ts:   ts,
		Noop: noopProvider,
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
	})
	if err != nil {
		return nil, err
	}
	c.Os = os

	ws, err := withdrawals.NewService(&withdrawals.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Fs:   fs,
		Ts:   ts,
		Noop: noopProvider,
	})
	if err != nil {
		return nil, err
	}
	c.WithdrawalService = ws

	ps, err := payments.NewService(&payments.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Ts:   ts,
		Noop: noopProvider,
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
		Noop:                noopProvider,
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

func (c *TestContainer) Cleanup(ctx context.Context) {
	c.Server.Close()
	c.Ctrl.Finish()
	c.Db.Close()
	c.Logger.Sync()
	c.Crdb.Container.Terminate(ctx)
}

func NewAccount(
	container *TestContainer,
	input *onboarding.CreateAccountArgs,
) (*accounts.Account, error) {
	container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
		Id: uuid.NewString(),
	}, nil).Times(1)

	acc, err := container.Os.CreateAccount(container.Ctx, input)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// Funcion will set the `AccountID` and `IdentityID` on the `verifyAccountArgs` for you.
func NewVerifiedAccount(
	container *TestContainer,
	createAccountArgs *onboarding.CreateAccountArgs,
	verifyAccountArgs *onboarding.VerifyAccountArgs,
) (*accounts.Account, error) {
	container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
		Id: uuid.NewString(),
	}, nil).Times(1)

	acc, err := container.Os.CreateAccount(container.Ctx, createAccountArgs)
	if err != nil {
		return nil, err
	}

	verifyAccountArgs.AccountID = acc.ID
	verifyAccountArgs.IdentityID = acc.IdentityID
	container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
			return &pacioliv1.Account{
				Id: args.Id,
			}, nil
		}).Times(2)
	acc, err = container.Os.VerifyAccount(container.Ctx, verifyAccountArgs)
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
	container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
			return &pacioliv1.Account{
				Id: args.Id,
			}, nil
		}).Times(1)
	bankAccount, err := container.FundingSourceService.CreateBankAccount(container.Ctx, args)
	if err != nil {
		return nil, err
	}

	if verify {
		container.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, args *pacioliv1.GetAccountRequest, opts ...grpc.CallOption) (*pacioliv1.Account, error) {
				return &pacioliv1.Account{
					Id: args.Id,
				}, nil
			}).Times(1)
		bankAccount, err = container.FundingSourceService.Verify(
			container.Ctx,
			&fundingsources.VerifyArgs{
				IdentityID:      user.ID,
				FundingSourceID: bankAccount.ID,
			})
	}

	return bankAccount, nil
}

func NewDeposit(
	c *TestContainer,
	user *_user.User,
	args *deposits.InitiateDepositArgs,
) (*account_transactions.AccountTransaction, error) {
	c.MockPacioliClient.EXPECT().GetAccount(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			args *pacioliv1.GetAccountRequest,
			opts ...grpc.CallOption,
		) (*pacioliv1.Account, error) {
			return &pacioliv1.Account{
				Id: args.Id,
			}, nil
		}).Times(2)
	c.MockPacioliClient.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).Return(
		&pacioliv1.CreateTransfersResponse{
			Errors: []*pacioliv1.EventError{},
		}, nil).Times(1)

	deposit, err := c.DepositService.InitiateDeposit(c.Ctx, args)
	if err != nil {
		return nil, err
	}

	return deposit, nil
}
