package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/interledger/interledger-app/go/backend/appcontext"
	"github.com/interledger/interledger-app/go/backend/errcodes"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

var errorStatus = map[error]struct {
	status int
	code   errcodes.AppErrorCode
}{
	user.ErrNoUserFound:            {http.StatusUnauthorized, errcodes.ErrCodeUserNoUserFound},
	wallets.ErrNoWalletFound:       {http.StatusNotFound, errcodes.ErrCodeWalletsNoWalletFound},
	wallets.ErrDuplicateWallet:     {http.StatusConflict, errcodes.ErrCodeWalletsDuplicateWallet},
	wallets.ErrWalletConflict:      {http.StatusConflict, errcodes.ErrCodeWalletsWalletConflict},
	linkedaccounts.ErrNotFound:     {http.StatusNotFound, errcodes.ErrCodeLinkedAccNotFound},
	kyc.ErrKYCResubmissionRequired: {http.StatusForbidden, errcodes.ErrCodeKYCResubmissionRequired},
	gatehub.ErrNotFound:            {http.StatusNotFound, errcodes.ErrCodeNotFound},
	gatehub.ErrInternal:            {http.StatusInternalServerError, errcodes.ErrCodeInternal},
	gatehub.ErrTimedOut:            {http.StatusGatewayTimeout, errcodes.ErrCodeGatewayTimeout},
	gatehub.ErrBadRequest:          {http.StatusBadRequest, errcodes.ErrCodeBadRequest},
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
	WriteAppError(w, r, http.StatusInternalServerError, errcodes.ErrCodeInternal, "Internal server error")
}

func WriteAppError(w http.ResponseWriter, r *http.Request, status int, code errcodes.AppErrorCode, message string) {
	reqID := appcontext.RequestIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(struct {
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
		ReqID     string `json:"req_id,omitempty"`
	}{code, message, reqID})
}
