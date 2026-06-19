package rpcserver

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/interledger/interledger-app/go/pacioli/db"
	"github.com/interledger/interledger-app/go/pacioli/healthcheck"
	test_utils "github.com/interledger/interledger-app/go/pacioli/utils"
	pacioliv1 "github.com/interledger/interledger-app/go/proto/pacioli/v1"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
)

type TestContainer struct {
	Ctx        context.Context
	Db         *sqlx.DB
	Server     *grpc.Server
	Client     pacioliv1.PacioliServiceClient
	Connection *grpc.ClientConn
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	_, c.Db = db.MigrateTestDB(t, ctx)

	b := test_utils.NewBackends(t, c.Db)

	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		return nil, err
	}

	hs, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	server := NewServer(b, hs)
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

	return nil
}
