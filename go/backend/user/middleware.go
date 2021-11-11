package user

import (
	"context"
	"fmt"
	"net/http"
)

func MakeMiddleware(us Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("cookie")

			// Don't allow unauthorised requests.
			if err != nil || c == nil {
				w.WriteHeader(401)
				fmt.Fprintf(w, "Unauthorized.")
				return
			}

			user, err := us.GetUser(c)
			if err != nil {
				http.Error(w, "Invalid cookie", http.StatusForbidden)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
