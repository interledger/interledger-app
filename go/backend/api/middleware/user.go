package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/interledger/interledger-app/go/backend/api/apperrors"
	"github.com/interledger/interledger-app/go/backend/errcodes"
	"github.com/interledger/interledger-app/go/backend/user"
)

func aalReason(err error) string {
	if errors.Is(err, user.ErrAAL1Required) {
		return "aal1_required"
	}
	if errors.Is(err, user.ErrAAL2Required) {
		return "aal2_required"
	}
	return ""
}

func MakeUserMiddleware(uc user.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				token = strings.TrimSpace(token)

				u, err := uc.UserForToken(ctx, token)
				if err != nil {
					if errors.Is(err, user.ErrAAL1Required) || errors.Is(err, user.ErrAAL2Required) {
						apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, aalReason(err))
						return
					}
					if !errors.Is(err, user.ErrNoUserFound) {
						apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "Error verifying bearer token.")
						return
					}
				}
				if u != nil {
					newCtx := context.WithValue(ctx, user.CtxKey, u)
					next.ServeHTTP(w, r.WithContext(newCtx))
					return
				}
			}

			cookie, err := r.Cookie("ory_kratos_session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			u, err := uc.UserForCookie(ctx, cookie.Value)
			if err != nil {
				if errors.Is(err, user.ErrAAL1Required) || errors.Is(err, user.ErrAAL2Required) {
					apperrors.WriteAppError(w, r, http.StatusUnauthorized, errcodes.ErrCodeUnauthorized, aalReason(err))
					return
				}
				if !errors.Is(err, user.ErrNoUserFound) {
					apperrors.WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "Error parsing session.")
					return
				}
			}

			if u == nil {
				next.ServeHTTP(w, r)
				return
			}

			newCtx := context.WithValue(ctx, user.CtxKey, u)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
