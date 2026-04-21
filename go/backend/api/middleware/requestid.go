package middleware

import (
	"context"
	"net/http"
)

type ctxKeyRequestIDType string

var ctxKeyRequestID = ctxKeyRequestIDType("request-id")

func MakeRequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			ctx := context.WithValue(r.Context(), ctxKeyRequestID, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	val := ctx.Value(ctxKeyRequestID)
	if val == nil {
		return ""
	}
	return val.(string)
}
