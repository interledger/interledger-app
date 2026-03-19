package grpc

import (
	"context"
	"time"

	user_middleware "gitlab.com/fynbos/backend/user/middleware"
	wallets_middleware "gitlab.com/fynbos/backend/wallets/middleware"
	"gitlab.com/fynbos/log"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	// "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type rpcService struct {
	b Backends
}

func NewServer(b Backends) (*grpc.Server, error) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryTimingInterceptor()),
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

// UnaryTimingInterceptor logs the duration of unary RPC calls.
func UnaryTimingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)

		log.Info("Unary RPC timing", zap.String("method", info.FullMethod), zap.Duration("duration", dur))

		return resp, err
	}
}
