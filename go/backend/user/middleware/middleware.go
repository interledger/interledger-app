package middleware

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cookieMetadataKey = "cookies"
)

// Our front-end will forward the raw http cookies in the metadata.
func MakeUnaryInterceptor(client user.Client) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Exclusion, services that do not require user/wallet auth.
		if strings.Contains(info.FullMethod, "BackendAdminService") {
			return handler(ctx, req)
		}

		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "Failed to parse metadata.")
		}

		authHeaders := meta.Get("Authorization")
		if len(authHeaders) > 0 {
			token := strings.TrimPrefix(authHeaders[0], "Bearer ")
			token = strings.TrimSpace(token)

			u, err := client.UserForToken(ctx, token)
			if err != nil {
				if !errors.Is(err, user.ErrNoUserFound) {
					return nil, status.Error(codes.Internal, "Error verifying bearer token.")
				}
			}
			if u != nil {
				newCtx := context.WithValue(ctx, user.CtxKey, u)
				return handler(newCtx, req)
			}
		}

		rawCookies := meta.Get(cookieMetadataKey) // must match the metadata field key set on the front-end

		if len(rawCookies) == 0 {
			return handler(ctx, req)
		}
		// cookies now contain the individual cookie names plus values in format <Name=Value>
		cookies := strings.Split(rawCookies[0], ";")

		var kratosCookie string
		for _, c := range cookies {
			k, v, _ := strings.Cut(c, "=")
			trimmed := strings.Trim(k, " ")
			if trimmed == "ory_kratos_session" {
				kratosCookie = v
			}
		}

		u, err := client.UserForCookie(ctx, kratosCookie)
		if err != nil {
			if !errors.Is(err, user.ErrNoUserFound) {
				return nil, status.Error(codes.Internal, "Error parsing session.")
			}
		}

		if u == nil {
			return handler(ctx, req)
		}

		newCtx := context.WithValue(ctx, user.CtxKey, u)

		return handler(newCtx, req)
	})
}
