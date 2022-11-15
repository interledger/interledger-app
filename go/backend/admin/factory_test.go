package admin_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/admin"
	"gitlab.com/fynbos/backend/openpayments"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/healthcheck"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/backend/waitlist"
	waitlist_mock "gitlab.com/fynbos/backend/waitlist/client/mock"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	Ctx           context.Context
	Ctrl          *gomock.Controller
	Db            *sqlx.DB
	Hs            healthcheck.Service
	Cs            country.Client
	Tp            *mocks.Client
	AdminConn     *grpc.ClientConn
	AdminClient   adminv1.BackendClient
	AdminServer   *grpc.Server
	ValidatorImpl *validator.Validate
	Auth          auth.Service
	WaitlistImpl  *waitlist_mock.MockClient
}

func (c *TestContainer) Users() user.Client {
	return nil
}

func (c *TestContainer) KYC() kyc.Client {
	return nil
}

func (c *TestContainer) OpenPayments() openpayments.Client {
	return nil
}

func (c *TestContainer) AdminAuth() auth.Service {
	return c.Auth
}

func (c *TestContainer) HealthCheck() healthcheck.Service {
	return c.Hs
}

func (c *TestContainer) Waitlist() waitlist.Client {
	return c.WaitlistImpl
}

func (c *TestContainer) Validator() *validator.Validate {
	return c.ValidatorImpl
}

func (c *TestContainer) DB() *sqlx.DB {
	return c.Db
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

	c.Ctrl = gomock.NewController(t)
	c.WaitlistImpl = waitlist_mock.NewMockClient(c.Ctrl)

	server, err := admin.NewServer(c)
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
	c.AdminClient = adminv1.NewBackendClient(adminConn)

	return c, nil
}
