package auth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MockService struct{}

func NewMockService() Service {
	return &MockService{}
}

func (m *MockService) ForContext(ctx context.Context) (*AdminUser, error) {
	raw, ok := ctx.Value(userCtxKey).(*AdminUser)
	if !ok || raw == nil {
		return nil, ErrNoUserFound
	}

	return raw, nil
}

func (m *MockService) GetAdminUser(ctx context.Context) (*AdminUser, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoUserFound
	}
	idTokens := meta.Get("idToken") // must match the metadata field key configured in Retool
	if len(idTokens) < 1 {
		return nil, ErrNoUserFound
	}
	if idTokens[0] == "" {
		return nil, ErrNoUserFound
	}

	return &AdminUser{Email: idTokens[0]}, nil
}

func (m *MockService) MakeUnaryInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !strings.Contains(info.FullMethod, "BackendAdminService") {
			return handler(ctx, req)
		}

		user, err := m.GetAdminUser(ctx)
		if err != nil {
			if !errors.Is(err, ErrNoUserFound) {
				return nil, status.Error(codes.Internal, "error parsing id token.")
			}
		}

		newCtx := ctx
		if user != nil {
			newCtx = context.WithValue(ctx, userCtxKey, user)
		}

		return handler(newCtx, req)
	})}
}

func ActingAs(ctx context.Context, email string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "idToken", email)
}
