package admin

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/noop"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/proto/backend/v1"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc"
)

type TestContainer struct {
	Ctx             context.Context
	Crdb            *test_utils.CockroachDBContainer
	Pacioli         *test_utils.PacioliContainer
	Db              *sqlx.DB
	As              accounts.Service
	Is              identity.Service
	Hs              healthcheck.Service
	Os              onboarding.Service
	Noop            noop.Service
	AdminConn       *grpc.ClientConn
	AdminClient     backend.BackendServiceClient
	AdminServer     *grpc.Server
	PacioliConn     *grpc.ClientConn
	PacioliClient   pacioliv1.PacioliServiceClient
	PacioliLedgerID uint16
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	if err := c.PacioliConn.Close(); err != nil {
		return err
	}

	if err := c.AdminConn.Close(); err != nil {
		return err
	}

	c.AdminServer.Stop()

	if err := c.Db.Close(); err != nil {
		return err
	}

	if err := c.Crdb.Container.Terminate(ctx); err != nil {
		return err
	}

	if err := c.Pacioli.Terminate(ctx); err != nil {
		return err
	}

	return nil
}

func NewTestContainer(ctx context.Context) (*TestContainer, error) {
	c := &TestContainer{}
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	pacioli, err := test_utils.SetupPacioli(ctx)
	if err != nil {
		return nil, err
	}
	c.Pacioli = pacioli

	pacioliConn, err := grpc.Dial(pacioli.PacioliUrl, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c.PacioliConn = pacioliConn
	c.PacioliClient = pacioliv1.NewPacioliServiceClient(pacioliConn)
	c.PacioliLedgerID = 1

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

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
	err = as.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.As = as

	equityAccID := uuid.NewString()
	noopProvider, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   equityAccID,
		PacioliTenant: "dev",
		PacioliClient: c.PacioliClient,
	})
	if err != nil {
		return nil, err
	}
	err = noopProvider.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.Noop = noopProvider

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

	listener, err := net.Listen("tcp", "0.0.0.0:8443")
	if err != nil {
		return nil, err
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	us := auth.NewMockService()
	server, err := NewServer(&ServerArgs{
		Hs: hs,
		Is: is,
		As: as,
		Us: us,
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

	adminConn, err := grpc.Dial("127.0.0.1:8443", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c.AdminConn = adminConn
	c.AdminClient = backend.NewBackendServiceClient(adminConn)

	return c, nil
}
