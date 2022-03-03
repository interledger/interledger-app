package user

import (
	"context"
	"errors"
	"net/http"
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
