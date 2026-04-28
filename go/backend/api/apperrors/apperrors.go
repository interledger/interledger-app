package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/backend/api/appcontext"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const (
	ErrCodeInternal       = "INTERNAL"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeGatewayTimeout = "GATEWAY_TIMEOUT"

	ErrCodeUserNoUserFound = "USER_NO_USER_FOUND"

	ErrCodeWalletsNoWalletFound   = "WALLETS_NO_WALLET_FOUND"
	ErrCodeWalletsDuplicateWallet = "WALLETS_DUPLICATE_WALLET"
	ErrCodeWalletsWalletConflict  = "WALLETS_WALLET_CONFLICT"

	ErrCodeLinkedAccNotFound = "LINKEDACC_NOT_FOUND"

	ErrCodeKYCResubmissionRequired = "KYC_RESUBMISSION_REQUIRED"
)

var errorStatus = map[error]struct {
	status int
	code   string
}{
	user.ErrNoUserFound:            {http.StatusUnauthorized, ErrCodeUserNoUserFound},
	wallets.ErrNoWalletFound:       {http.StatusNotFound, ErrCodeWalletsNoWalletFound},
	wallets.ErrDuplicateWallet:     {http.StatusConflict, ErrCodeWalletsDuplicateWallet},
	wallets.ErrWalletConflict:      {http.StatusConflict, ErrCodeWalletsWalletConflict},
	linkedaccounts.ErrNotFound:     {http.StatusNotFound, ErrCodeLinkedAccNotFound},
	kyc.ErrKYCResubmissionRequired: {http.StatusForbidden, ErrCodeKYCResubmissionRequired},
	gatehub.ErrNotFound:            {http.StatusNotFound, ErrCodeNotFound},
	gatehub.ErrInternal:            {http.StatusInternalServerError, ErrCodeInternal},
	gatehub.ErrTimedOut:            {http.StatusGatewayTimeout, ErrCodeGatewayTimeout},
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
	WriteAppError(w, r, http.StatusInternalServerError, ErrCodeInternal, "Internal server error")
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
