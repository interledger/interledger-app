package grpc

import (
	_admin "gitlab.com/fynbos/backend/admin"
	open_server "gitlab.com/fynbos/backend/openpayments/server"
	user_middleware "gitlab.com/fynbos/backend/user/middleware"
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
		b.AdminAuth().MakeUnaryInterceptors(),
		user_middleware.MakeUnaryInterceptor(b.Users()),
	)
	backendv1.RegisterBackendServiceServer(server, &rpcService{
		b: b,
	})
	backendv1.RegisterBackendAdminServiceServer(server, &_admin.AdminRpcService{
		Validator:   b.Validator(),
		AuthService: b.AdminAuth(),
		Temporal:    b.Temporal(),
	})
	backendv1.RegisterOpenPaymentServiceServer(server, open_server.NewGRPCServer(b))
	grpc_health_v1.RegisterHealthServer(server, b.HealthCheck())
	reflection.Register(server)
	return server, nil
}
