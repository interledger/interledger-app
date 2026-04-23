package middleware

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cookieMetadataKey = "cookies"
)

func aalErrorToStatus(err error) error {
	var reason string
	var appErrorCode pb.ErrorCode
	if errors.Is(err, user.ErrAAL1Required) {
		reason = "aal1_required"
		appErrorCode = pb.ErrorCode_ERROR_CODE_USER_AAL1_REQUIRED
	} else if errors.Is(err, user.ErrAAL2Required) {
		reason = "aal2_required"
		appErrorCode = pb.ErrorCode_ERROR_CODE_USER_AAL2_REQUIRED
	} else {
		return nil
	}

	st := status.New(codes.Unauthenticated, "Unauthenticated")
	metadata := &errdetails.ErrorInfo{
		Reason: reason,
		Domain: "authentication",
	}

	appError := &pb.AppError{
		ErrorCode: appErrorCode.String(),
		Message:   reason,
	}

	st, detailErr := st.WithDetails(metadata, appError)
	if detailErr != nil {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}
	return st.Err()
}

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
				// Return AAL errors immediately with metadata
				if errors.Is(err, user.ErrAAL1Required) || errors.Is(err, user.ErrAAL2Required) {
					return nil, aalErrorToStatus(err)
				}
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
			// Return AAL errors immediately with metadata
			if errors.Is(err, user.ErrAAL1Required) || errors.Is(err, user.ErrAAL2Required) {
				return nil, aalErrorToStatus(err)
			}
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
