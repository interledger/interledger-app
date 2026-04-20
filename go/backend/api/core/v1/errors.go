package v1

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var errorStatus = map[error]int{
	gatehub.ErrNotFound: http.StatusForbidden,
}

func toHTTPError(w http.ResponseWriter, err error) {
	log.Info("http error", zap.Error(err))

	for k, status := range errorStatus {
		if errors.Is(err, k) {
			http.Error(w, http.StatusText(status), status)
			return
		}
	}

	_ = sentry.CaptureException(err)
	log.Error("unexpected error", zap.Error(err))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
