package middleware

import (
	"net/http"

	"gitlab.com/fynbos/backend/api/appcontext"
)

func MakeRequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			ctx := appcontext.WithRequestID(r.Context(), reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
