package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/interledger/interledger-app/go/backend/api/apperrors"
	"github.com/interledger/interledger-app/go/backend/errcodes"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

func MakeWalletMiddleware(uc user.Client, wc wallets.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			u, err := uc.UserForContext(ctx)
			if err != nil && !errors.Is(err, user.ErrNoUserFound) {
				apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "Error parsing session.")
				return
			}

			if u == nil {
				next.ServeHTTP(w, r)
				return
			}

			walletList, err := wc.List(ctx, u.ID)
			if err != nil {
				// Do nothing for now.
				next.ServeHTTP(w, r)
				return
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
					next.ServeHTTP(w, r)
					return
				}
			}

			if len(walletList) > 1 {
				log.Warn("user has multiple wallets, using a default", zap.String("user_id", u.ID))
			}

			newCtx := context.WithValue(ctx, wallets.CtxKey, &walletList[0])
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
