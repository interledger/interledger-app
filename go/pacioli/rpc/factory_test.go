package rpc

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli/healthcheck"
	"gitlab.com/fynbos/pacioli/ledger"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc"
)

type TestContainer struct {
	TbClient   tigerbeetle_go.Client
	Ctx        context.Context
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

	_, c.Db = test_utils.MigrateCockroachDB(t, ctx)

	tb, err := test_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		return nil, err
	}
	c.Tb = tb

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tb.URI}, 1000)
	if err != nil {
		return nil, err
	}
	c.TbClient = tbClient

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: c.Db,
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
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	c.TbClient.Close()

	err := c.Tb.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}
