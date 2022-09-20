package admin_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	user_mock "gitlab.com/fynbos/backend/user/client/mock"

	agreements_client "gitlab.com/fynbos/backend/agreements/client"

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
	"gitlab.com/fynbos/backend/fundingsources"
	funding_mock "gitlab.com/fynbos/backend/fundingsources/client/mock"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	identity_client "gitlab.com/fynbos/backend/identity/client"
	"gitlab.com/fynbos/backend/onboarding"
	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"
	"gitlab.com/fynbos/backend/providers/mx"
	mx_mock "gitlab.com/fynbos/backend/providers/mx/client/mock"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/supporttickets"
	support_mock "gitlab.com/fynbos/backend/supporttickets/client/mock"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/backend/waitlist"
	waitlist_mock "gitlab.com/fynbos/backend/waitlist/client/mock"
	"gitlab.com/fynbos/pacioli"
	"gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/client"
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
	Cs              country.Client
	Tp              *mocks.Client
	RafikiProvider  *rafiki.MockService
	AdminConn       *grpc.ClientConn
	AdminClient     backend.BackendAdminServiceClient
	AdminServer     *grpc.Server
	PacioliClient   pacioli.Client
	PacioliLedgerID uint32
	ValidatorImpl   *validator.Validate
	Auth            auth.Service
	AgreementsImpl  agreements.Client
	Mx              *mx_mock.MockClient
	Tickets         *support_mock.MockClient
	UsersImpl       user.Client
	WaitlistImpl    *waitlist_mock.MockClient
	TwilioImpl      *twilio.MockService
}

func (c *TestContainer) AdminAuth() auth.Service {
	return c.Auth
}

func (c *TestContainer) Agreements() agreements.Client {
	return c.AgreementsImpl
}

func (c *TestContainer) FundingSources() fundingsources.Client {
	return c.Fs
}

func (c *TestContainer) HealthCheck() healthcheck.Service {
	return c.Hs
}

func (c *TestContainer) MX() mx.Client {
	return c.Mx
}

func (c *TestContainer) Onboarding() onboarding.Client {
	return c.Os
}

func (c *TestContainer) Rafiki() rafiki.Service {
	return c.RafikiProvider
}

func (c *TestContainer) SupportTickets() supporttickets.Client {
	return c.Tickets
}

func (c *TestContainer) Twilio() twilio.Service {
	return c.TwilioImpl
}

func (c *TestContainer) Unit() unit.Client {
	return c.Up
}

func (c *TestContainer) Waitlist() waitlist.Client {
	return c.WaitlistImpl
}

func (c *TestContainer) Accounts() accounts.Client {
	return c.As
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
	c.Hs = hs

	att := auth.NewMockService()
	c.Auth = att

	c.AgreementsImpl = agreements_client.New(c)
	c.Ctrl = gomock.NewController(t)
	c.Fs = funding_mock.NewMockClient(c.Ctrl)
	c.RafikiProvider = rafiki.NewMockService(c.Ctrl)
	c.TwilioImpl = twilio.NewMockService(c.Ctrl)
	c.UsersImpl = user_mock.NewMock()
	c.Mx = mx_mock.NewMockClient(c.Ctrl)
	c.WaitlistImpl = waitlist_mock.NewMockClient(c.Ctrl)
	c.Tickets = support_mock.NewMockClient(c.Ctrl)

	server, err := _grpc.NewServer(c)
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
