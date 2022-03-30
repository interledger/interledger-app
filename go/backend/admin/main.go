package admin

import (
	"context"
	"log"

	"gitlab.com/fynbos/backend/healthcheck"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(health healthcheck.Service) *grpc.Server {
	server := grpc.NewServer()
	backendv1.RegisterBackendServiceServer(server, &rpcServer{})
	grpc_health_v1.RegisterHealthServer(server, health)
	reflection.Register(server)
	return server
}

type rpcServer struct {
	backendv1.UnimplementedBackendServiceServer
}

func (s *rpcServer) GetUserAccountByEmail(
	ctx context.Context,
	req *backendv1.GetUserAccountByEmailRequest,
) (*backendv1.Account, error) {
	log.Println("admin getting account. email: ", req.GetEmail())
	return &backendv1.Account{
		Id:      "1234",
		Balance: 100,
	}, nil
}
