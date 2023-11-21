package middleware

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/wallets"

	"github.com/gogo/status"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Our front-end will forward the raw http cookies in the metadata.
func MakeUnaryInterceptor(uc user.Client, wc wallets.Client) grpc.ServerOption {
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

		u, err := uc.UserForContext(ctx)
		if err != nil && !errors.Is(err, user.ErrNoUserFound) {
			return nil, status.Error(codes.Internal, "Error parsing session.")
		}

		if u == nil {
			return handler(ctx, req)
		}

		walletList, err := wc.List(ctx, u.ID)
		if err != nil {
			// Do nothing for now.
			return handler(ctx, req)
		}

		// Create a default wallet for the user if they don't already have one
		if len(walletList) == 0 {
			_, err = wc.Create(ctx, wallets.CreateArgs{
				UserID:  u.ID,
				Country: u.Country,
			})
			if err != nil && !errors.Is(err, wallets.ErrDuplicateWallet) {
				log.Warn("failed to create default wallet for user", zap.Error(err), zap.String("user_id", u.ID))
			}
			walletList, err = wc.List(ctx, u.ID)
			if err != nil || len(walletList) <= 0 {
				// Do nothing for now. We tried and the next request will try again
				return handler(ctx, req)
			}
		}

		if len(walletList) > 1 {
			log.Warn("user has multiple wallets, using a default", zap.String("user_id", u.ID))
		}

		newCtx := context.WithValue(ctx, wallets.CtxKey, &walletList[0])

		return handler(newCtx, req)
	})
}
