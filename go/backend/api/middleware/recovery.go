package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type responseTracker struct {
	http.ResponseWriter
	started bool
}

func (rt *responseTracker) WriteHeader(code int) {
	rt.started = true
	rt.ResponseWriter.WriteHeader(code)
}

func (rt *responseTracker) Write(b []byte) (int, error) {
	rt.started = true
	return rt.ResponseWriter.Write(b)
}

func MakeRecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rt := &responseTracker{ResponseWriter: w}

			defer func() {
				if rec := recover(); rec != nil {
					err := fmt.Errorf("panic: %v", rec)
					_ = sentry.CaptureException(err)
					log.Error("panic recovered in HTTP handler", zap.Any("panic", rec))

					if !rt.started {
						WriteAppError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Internal server error")
					}
				}
			}()

			next.ServeHTTP(rt, r)
		})
	}
}
