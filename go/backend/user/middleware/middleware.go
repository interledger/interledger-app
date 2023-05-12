package middleware

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"gitlab.com/fynbos/log"

	"gitlab.com/fynbos/backend/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	cookieMetadataKey = "cookies"
	userCtxKey        = user.UserCtxKey("user")
	walletCtxKey      = user.WalletCtxKey("wallet")
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

		newCtx := context.WithValue(ctx, userCtxKey, u)

		wallets, err := client.ListWallets(newCtx, u.ID)
		if err != nil {
			// Do nothing for now.
			return handler(newCtx, req)
		}

		// Create a default wallet for the user if they don't already have one
		if len(wallets) == 0 {
			_, err = client.CreateNewWallet(ctx, user.CreateWalletArgs{
				UserID: u.ID,
			})
			if err != nil && !errors.Is(err, user.ErrDuplicateWallet) {
				log.Warn("failed to create default wallet for user", zap.Error(err), zap.String("user_id", u.ID))
			}
			wallets, err = client.ListWallets(newCtx, u.ID)
			if err != nil || len(wallets) <= 0 {
				// Do nothing for now. We tried and the next request will try again
				return handler(newCtx, req)
			}
		}

		if len(wallets) > 1 {
			log.Warn("user has multiple wallets, using a default", zap.String("user_id", u.ID))
		}

		newCtx = context.WithValue(newCtx, walletCtxKey, &wallets[0])

		return handler(newCtx, req)
	})
}
