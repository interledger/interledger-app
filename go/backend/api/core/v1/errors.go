package v1

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	api_middleware "gitlab.com/fynbos/backend/api/middleware"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var errorStatus = map[error]struct {
	status int
	code   string
}{
	user.ErrNoUserFound:            {http.StatusUnauthorized, api_middleware.ErrCodeUserNoUserFound},
	wallets.ErrNoWalletFound:       {http.StatusNotFound, api_middleware.ErrCodeWalletsNoWalletFound},
	wallets.ErrDuplicateWallet:     {http.StatusConflict, api_middleware.ErrCodeWalletsDuplicateWallet},
	wallets.ErrWalletConflict:      {http.StatusConflict, api_middleware.ErrCodeWalletsWalletConflict},
	linkedaccounts.ErrNotFound:     {http.StatusNotFound, api_middleware.ErrCodeLinkedAccNotFound},
	kyc.ErrKYCResubmissionRequired: {http.StatusForbidden, api_middleware.ErrCodeKYCResubmissionRequired},
	gatehub.ErrNotFound:            {http.StatusNotFound, api_middleware.ErrCodeGatehubNotFound},
	gatehub.ErrInternal:            {http.StatusInternalServerError, api_middleware.ErrCodeGatehubInternal},
	gatehub.ErrTimedOut:            {http.StatusGatewayTimeout, api_middleware.ErrCodeGatehubTimedOut},
}

func toHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	if !env.IsTest() {
		log.Info("http error", zap.Error(err))
	}

	if v, ok := errorStatus[err]; ok {
		api_middleware.WriteAppError(w, r, v.status, v.code, http.StatusText(v.status))
		return
	}

	for k, v := range errorStatus {
		if errors.Is(err, k) {
			api_middleware.WriteAppError(w, r, v.status, v.code, http.StatusText(v.status))
			return
		}
	}

	_ = sentry.CaptureException(err)
	log.Error("unexpected error", zap.Error(err))
	api_middleware.WriteAppError(w, r, http.StatusInternalServerError, api_middleware.ErrCodeInternal, "Internal server error")
}
