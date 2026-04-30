package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/backend/appcontext"
	"gitlab.com/fynbos/backend/errorhandling"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var errorStatus = map[error]struct {
	status int
	code   string
}{
	user.ErrNoUserFound:            {http.StatusUnauthorized, errorhandling.ErrCodeUserNoUserFound},
	wallets.ErrNoWalletFound:       {http.StatusNotFound, errorhandling.ErrCodeWalletsNoWalletFound},
	wallets.ErrDuplicateWallet:     {http.StatusConflict, errorhandling.ErrCodeWalletsDuplicateWallet},
	wallets.ErrWalletConflict:      {http.StatusConflict, errorhandling.ErrCodeWalletsWalletConflict},
	linkedaccounts.ErrNotFound:     {http.StatusNotFound, errorhandling.ErrCodeLinkedAccNotFound},
	kyc.ErrKYCResubmissionRequired: {http.StatusForbidden, errorhandling.ErrCodeKYCResubmissionRequired},
	gatehub.ErrNotFound:            {http.StatusNotFound, errorhandling.ErrCodeNotFound},
	gatehub.ErrInternal:            {http.StatusInternalServerError, errorhandling.ErrCodeInternal},
	gatehub.ErrTimedOut:            {http.StatusGatewayTimeout, errorhandling.ErrCodeGatewayTimeout},
}

func ToHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	log.Info("http error", zap.Error(err))

	if v, ok := errorStatus[err]; ok {
		WriteAppError(w, r, v.status, v.code, http.StatusText(v.status))
		return
	}

	for k, v := range errorStatus {
		if errors.Is(err, k) {
			WriteAppError(w, r, v.status, v.code, http.StatusText(v.status))
			return
		}
	}

	sentry.CaptureException(err)
	log.Error("unexpected error", zap.Error(err))
	WriteAppError(w, r, http.StatusInternalServerError, errorhandling.ErrCodeInternal, "Internal server error")
}

func WriteAppError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := appcontext.RequestIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(struct {
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
		ReqID     string `json:"req_id,omitempty"`
	}{code, message, reqID})
}
