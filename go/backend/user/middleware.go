package user

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func MakeMiddleware(us Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := us.GetUser(*r)
			if err != nil {
				if !errors.Is(err, ErrNoUserFound) {
					http.Error(w, "error parsing session", http.StatusInternalServerError)
					return
				}
			}

			// If no user don't add to context
			if user != nil {
				// put it in context
				ctx := context.WithValue(r.Context(), userCtxKey, user)

				// and call the next with our new context
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Our front-end will forward the raw http cookies in the metadata.
func MakeUnaryInterceptor(us Service, serviceName string) grpc.ServerOption {
	type grpcService interface {
		GetName() string
	}
	return grpc.ChainUnaryInterceptor(func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		service := info.Server.(grpcService)
		if service.GetName() != serviceName {
			return handler(ctx, req)
		}

		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "Failed to parse metadata.")
		}
		rawCookies := meta.Get("cookies") // must match the metadata field key set on the front-end
		r := http.Request{}               // use std lib to parse cookies
		if len(rawCookies) > 1 {
			r.Header.Set("Cookie", rawCookies[0]) // we assume there'll be one
		}

		user, err := us.GetUser(r)
		if err != nil {
			if !errors.Is(err, ErrNoUserFound) {
				return nil, status.Error(codes.Internal, "Error parsing session.")
			}
		}

		newCtx := ctx
		// If no user don't add to context
		if user != nil {
			// put it in context
			newCtx = context.WithValue(r.Context(), userCtxKey, user)
		}

		return handler(newCtx, req)
	})
}
