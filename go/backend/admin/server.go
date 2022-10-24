package admin

import (
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(b Backends) (*grpc.Server, error) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		b.AdminAuth().MakeUnaryInterceptors(),
	)

	adminv1.RegisterBackendServer(server, &AdminRpcService{
		b: b,
	})

	reflection.Register(server)
	return server, nil
}
