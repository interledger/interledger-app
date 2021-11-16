package user

import (
	"context"
	"net/http"
)

func MakeMiddleware(us Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := us.GetUser(*r)
			if err != nil {
				switch err.(type) {
				case *NoCookieError:
					http.Error(w, "Unauthorized.", http.StatusUnauthorized)
					return
				default:
					http.Error(w, "Invalid cookie.", http.StatusForbidden)
					return
				}
			}

			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
