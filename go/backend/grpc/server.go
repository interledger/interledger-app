package grpc

import (
	"errors"

	_admin "gitlab.com/fynbos/backend/admin"
	user_middleware "gitlab.com/fynbos/backend/user/middleware"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	ErrInvalidArgument = errors.New("grpc: invalid argument")
	ErrInternal        = errors.New("grpc: internal error")
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
		Validator:       b.Validator(),
		AccountsService: b.Accounts(),
		IdentityService: b.Identity(),
		AuthService:     b.AdminAuth(),
		Temporal:        b.Temporal(),
	})
	grpc_health_v1.RegisterHealthServer(server, b.HealthCheck())
	reflection.Register(server)
	return server, nil
}
