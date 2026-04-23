package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MakeUnaryInterceptorRequestId gets the "x-request-id" value from the incoming
// request metadata and adds it to the context for easy access.
func MakeUnaryInterceptorRequestId() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "Failed to parse metadata.")
		}

		ctx, err := ctxWithRequestIdFromMeta(ctx, meta)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result, err := handler(ctx, req)

		return result, err
	})
}

func ctxWithRequestIdFromMeta(ctx context.Context, meta metadata.MD) (context.Context, error) {
	reqId, err := RequestIdFromMetadata(meta)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, ctxKeyRequestId, reqId)
	return ctx, nil
}
