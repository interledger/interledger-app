package rpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli/healthcheck"
	"gitlab.com/fynbos/pacioli/ledger"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"gitlab.com/fynbos/tigerbeetle_go"
	"google.golang.org/grpc"
)

type TestContainer struct {
	TbClient   tigerbeetle_go.Client
	Ctx        context.Context
	Crdb       *test_utils.CockroachDBContainer
	Tb         *test_utils.TigerBeetleContainer
	Db         *sqlx.DB
	Ls         ledger.Service
	Server     *grpc.Server
	Client     pacioliv1.PacioliServiceClient
	Connection *grpc.ClientConn
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	containerNetwork := "pacioli-test"

	crdb, err := test_utils.SetupTestCockroachDB(ctx, containerNetwork)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

	tb, err := test_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		return nil, err
	}
	c.Tb = tb

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tb.URI})
	if err != nil {
		return nil, err
	}
	// drive the TB client.
	go func() {
		tick := time.Tick(20 * time.Millisecond)
		for range tick {
			tbClient.Tick()
		}
	}()

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: db,
		Tb: tbClient,
	})
	if err != nil {
		return nil, err
	}
	c.Ls = ls

	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		return nil, err
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	server := NewServer(ls, hs)
	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()
	c.Server = server

	connectTo := "127.0.0.1:8081"
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c.Connection = conn
	client := pacioliv1.NewPacioliServiceClient(conn)
	c.Client = client

	return c, nil
}

func (c *TestContainer) Cleanup() error {
	c.Server.Stop()
	err := c.Db.Close()
	if err != nil {
		return err
	}

	err = c.Crdb.Container.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	err = c.Tb.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}
