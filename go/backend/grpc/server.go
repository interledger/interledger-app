package grpc

import (
	user_middleware "gitlab.com/fynbos/backend/user/middleware"
	wallets_middleware "gitlab.com/fynbos/backend/wallets/middleware"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type rpcService struct {
	b Backends
}

func NewServer(b Backends) (*grpc.Server, error) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		MakeUnaryInterceptorRequestId(),
		MakeUnaryInterceptorAppError(),
		user_middleware.MakeUnaryInterceptor(b.Users()),
		wallets_middleware.MakeUnaryInterceptor(b.Users(), b.Wallets()),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcService{
		b: b,
	})

	grpc_health_v1.RegisterHealthServer(server, b.HealthCheck())
	reflection.Register(server)
	return server, nil
}
