package admin

import (
	"gitlab.com/fynbos/backend/healthcheck"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(b Backends) (*grpc.Server, error) {
	interceptors := append([]grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor())},
		b.AdminAuth().MakeUnaryInterceptors()...)
	server := grpc.NewServer(
		interceptors...,
	)

	adminv1.RegisterBackendServer(server, &AdminRpcService{
		b: b,
	})

	health, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}
	grpc_health_v1.RegisterHealthServer(server, health)

	reflection.Register(server)
	return server, nil
}
