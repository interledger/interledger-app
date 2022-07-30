package admin_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/osohq/go-oso"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/backend/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc"
)

type TestContainer struct {
	Ctx             context.Context
	Ctrl            *gomock.Controller
	Pacioli         *test_utils.PacioliContainer
	Db              *sqlx.DB
	As              accounts.Service
	Fs              *fundingsources.MockService
	Is              identity.Service
	Hs              healthcheck.Service
	Os              onboarding.Service
	Oso             *oso.Oso
	Noop            noop.Service
	Up              unit.Service
	Tp              *mocks.Client
	RafikiProvider  *rafiki.MockService
	AdminConn       *grpc.ClientConn
	AdminClient     backend.BackendAdminServiceClient
	AdminServer     *grpc.Server
	PacioliConn     *grpc.ClientConn
	PacioliClient   pacioliv1.PacioliServiceClient
	PacioliLedgerID uint32
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	if err := c.PacioliConn.Close(); err != nil {
		return err
	}

	if err := c.AdminConn.Close(); err != nil {
		return err
	}

	c.AdminServer.Stop()

	return nil
}

// TODO: refactor how we spin up deps for tests.
func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	db := test_utils.MigrateCockroachDB(t, ctx)
	c.Db = db

	c.Pacioli = test_utils.SetupPacioli(t, ctx)

	pacioliConn, err := grpc.Dial(c.Pacioli.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c.PacioliConn = pacioliConn
	c.PacioliClient = pacioliv1.NewPacioliServiceClient(pacioliConn)
	c.PacioliLedgerID = 1

	cs := country.NewService(db)
	is, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
	c.Is = is

	as, err := accounts.NewService(&accounts.ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: c.PacioliLedgerID,
		PacioliClient:   c.PacioliClient,
		Db:              db,
	})
	if err != nil {
		return nil, err
	}

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

	c.Noop = noopProvider
	c.Tp = &mocks.Client{}
	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   is,
		Noop: noopProvider,
		Tp:   c.Tp,
	})
	if err != nil {
		return nil, err
	}
	c.Os = os

	listener, err := net.Listen("tcp", "0.0.0.0:8443")
	if err != nil {
		return nil, err
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	us := auth.NewMockService()

	ags, err := agreements.NewService(&agreements.ServiceArgs{
		Db:            db,
		AgreementsDir: "../utils/agreements/test",
	})
	if err != nil {
		return nil, err
	}

	c.Ctrl = gomock.NewController(t)
	c.Up = unit.NewMockService(c.Ctrl)
	c.Fs = fundingsources.NewMockService(c.Ctrl)
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

	adminConn, err := grpc.Dial("127.0.0.1:8443", grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
