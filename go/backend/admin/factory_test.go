package admin_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	support_mock "gitlab.com/fynbos/backend/supporttickets/client/mock"
	waitlist_mock "gitlab.com/fynbos/backend/waitlist/client/mock"

	"go.temporal.io/sdk/client"

	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"

	identity_client "gitlab.com/fynbos/backend/identity/client"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/osohq/go-oso"
	"gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/country"
	country_client "gitlab.com/fynbos/backend/country/client"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	funding_mock "gitlab.com/fynbos/backend/fundingsources/client/mock"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	unit_mock "gitlab.com/fynbos/backend/providers/unit/client/mock"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/pacioli"
	"gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	Ctx             context.Context
	Ctrl            *gomock.Controller
	Db              *sqlx.DB
	As              accounts.Client
	Fs              fundingsources.Client
	Is              identity.Client
	Hs              healthcheck.Service
	Os              onboarding.Client
	Oso             *oso.Oso
	NoopImpl        noop.Service
	Up              unit.Client
	Cs              country.Client
	Tp              *mocks.Client
	RafikiProvider  *rafiki.MockService
	AdminConn       *grpc.ClientConn
	AdminClient     backend.BackendAdminServiceClient
	AdminServer     *grpc.Server
	PacioliClient   pacioli.Client
	PacioliLedgerID uint32
	ValidatorImpl   *validator.Validate
}

func (c *TestContainer) Accounts() accounts.Client {
	return c.As
}

func (c *TestContainer) Noop() noop.Service {
	return c.NoopImpl
}

func (c *TestContainer) Temporal() client.Client {
	return c.Tp
}

func (c *TestContainer) Validator() *validator.Validate {
	return c.ValidatorImpl
}

func (c *TestContainer) DB() *sqlx.DB {
	return c.Db
}

func (c *TestContainer) Identity() identity.Client {
	return c.Is
}

func (c *TestContainer) Countries() country.Client {
	return c.Cs
}

func (c *TestContainer) Pacioli() pacioli.Client {
	return c.PacioliClient
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	if err := c.AdminConn.Close(); err != nil {
		return err
	}

	c.AdminServer.Stop()

	return nil
}

// TODO: refactor how we spin up deps for tests.
func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{
		ValidatorImpl: validator.New(),
	}
	db := test_utils.MigrateCockroachDB(t, ctx)
	c.Db = db

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	pClient := test_utils.SetupPacioli(t, ctx)
	c.PacioliClient = pClient
	c.PacioliLedgerID = 1

	cs := country_client.New(c)
	c.Cs = cs

	is := identity_client.New(c, logger)
	c.Is = is

	as := accounts_client.New(c, c.PacioliLedgerID, logger)

	c.As = as

	equityAccID := "46d4b2bd-e29b-4a63-9aa8-7990776c714e"
	noopProvider, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   equityAccID,
		PacioliTenant: "dev",
		PacioliClient: c.PacioliClient,
	})
	if err != nil {
		return nil, err
	}

	c.NoopImpl = noopProvider
	c.Tp = &mocks.Client{}

	os := onboarding_client.New(c)
	c.Os = os
	port, err := test_utils.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	us := auth.NewMockService()

	ags, err := agreements.NewService(&agreements.ServiceArgs{
		Db: db,
	})
	if err != nil {
		return nil, err
	}

	c.Ctrl = gomock.NewController(t)
	c.Up = unit_mock.NewMockClient(c.Ctrl)
	c.Fs = funding_mock.NewMockClient(c.Ctrl)
	c.RafikiProvider = rafiki.NewMockService(c.Ctrl)
	tw := twilio.NewMockService(c.Ctrl)
	server, err := _grpc.NewServer(&_grpc.ServerArgs{
		HealthCheckService:   hs,
		IdentityService:      is,
		AccountsService:      as,
		AdminAuthService:     us,
		AgreementsService:    ags,
		UnitProvider:         c.Up,
		UserService:          user.NewMockService(),
		FundingSourceService: c.Fs,
		TwilioService:        tw,
		OnboardingService:    os,
		MxProvider:           mx.NewMockService(c.Ctrl),
		RafikiProvider:       c.RafikiProvider,
		DepositService:       deposits.NewMockService(c.Ctrl),
		WaitlistClient:       waitlist_mock.NewMockClient(c.Ctrl),
		Temporal:             c.Tp,
		TicketClient:         support_mock.NewMockClient(c.Ctrl),
		PaymentsClient:       payments_mock.NewMockClient(c.Ctrl),
	})
	if err != nil {
		return nil, err
	}
	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()
	c.AdminServer = server

	adminConn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c.AdminConn = adminConn
	c.AdminClient = backend.NewBackendAdminServiceClient(adminConn)

	return c, nil
}

// note https://github.com/golang/mock/issues/139
// we run the grpc server in a goroutine so any errors from the mock will be swallowed.
// TODO: see if there is a test server for grpc
func NewAdminClient(t *testing.T, svr *grpc.Server) (client backend.BackendAdminServiceClient, cleanup func()) {
	port, err := test_utils.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		t.Fatal(err)
	}

	// run the server
	go func() {
		if err := svr.Serve(listener); err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", port), grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	client = backend.NewBackendAdminServiceClient(conn)
	cleanup = func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}

	return client, cleanup
}
