package graph

import (
	"context"
	"errors"
	"net/http/httptest"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/accounts"
	_account "gitlab.com/fynbos/backend/accounts"
	_country "gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity/noop"
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
	NoopProvider         *noop.MockProvider
	TransactionService   account_transactions.Service
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

	provider := noop.NewMockProvider(ctrl)
	c.NoopProvider = provider

	is, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
		NoopProvider:   provider,
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

	as, err := accounts.NewService(is, cs, ledgerCode, pClient)
	if err != nil {
		return nil, err
	}
	c.AccountService = accounts.NewLoggingService(as, logger)

	users := _user.NewMockService()
	c.UserService = _user.NewLoggingService(users, logger)

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Identity: is,
	})
	if err != nil {
		return nil, err
	}
	c.FundingSourceService = fundingsources.NewLoggingService(fs, logger)

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
		Db:            db,
		FundingSource: fs,
		Account:       as,
		Transaction:   ts,
		LedgerID:      ledgerID,
		EquityAccID:   equityAccID,
	})
	if err != nil {
		return nil, err
	}
	c.NoopService = noopProvider

	graph, err := NewService(GraphqlOpts{
		Db:       db,
		Identity: is,
		User:     users,
		Account:  as,
		Noop:     noopProvider,
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

func NewIdentity(
	container *TestContainer,
	user *_user.User,
	input *generated.CreateIdentityInput,
) (*identity.Identity, error) {
	container.MockPacioliClient.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&pacioliv1.Account{
		Id: uuid.NewString(),
	}, nil).Times(1)
	req := createIdentityRequest(input)
	_user.ActingAs(req, user)
	var data map[string]generated.CreateIdentityMutationResponse
	if err := container.Client.Run(container.Ctx, req, &data); err != nil {
		return nil, err
	}

	response := data["createIdentity"]
	if response.Code != "200" && response.Success {
		return nil, errors.New(response.Message)
	}

	return response.Identity, nil
}

func NewLinkedUsdBankAccount(
	container *TestContainer,
	user *_user.User,
	input *generated.LinkUsdBankAccountInput,
	verifyFundingSource bool,
) (*generated.FundingSource, error) {
	req := linkUsdBankAccountRequest(input)
	_user.ActingAs(req, user)
	var linkingData map[string]generated.LinkFundingSourceMutationResponse
	if err := container.Client.Run(container.Ctx, req, &linkingData); err != nil {
		return nil, err
	}

	response := linkingData["linkUsdBankAccount"]
	if response.Code != "200" && response.Success {
		return nil, errors.New(response.Message)
	}

	if verifyFundingSource {
		verifyReq := verifyUsdBankAccount(
			generateVerifyUsdBankAccountInput(withFundingSourceID(response.FundingSource.ID)),
		)
		_user.ActingAs(verifyReq, user)
		var verifyData map[string]generated.VerifyMutationResponse
		if err := container.Client.Run(container.Ctx, verifyReq, &verifyData); err != nil {
			return nil, errors.New(response.Message)
		}
	}

	return response.FundingSource, nil
}
