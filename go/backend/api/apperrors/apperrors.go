package apperrors

import (
	"encoding/json"
	"net/http"

	"github.com/getsentry/sentry-go"
	"gitlab.com/fynbos/backend/appcontext"
	"gitlab.com/fynbos/backend/errorhandling"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var errorStatus = map[errorhandling.AppErrorCode]int{
	errorhandling.ErrCodeInternal:                  http.StatusInternalServerError,
	errorhandling.ErrCodeUnauthorized:              http.StatusUnauthorized,
	errorhandling.ErrCodeNotFound:                  http.StatusNotFound,
	errorhandling.ErrCodeConflict:                  http.StatusConflict,
	errorhandling.ErrCodeForbidden:                 http.StatusForbidden,
	errorhandling.ErrCodeBadRequest:                http.StatusBadRequest,
	errorhandling.ErrCodeGatewayTimeout:            http.StatusGatewayTimeout,
	errorhandling.ErrCodeValidation:                http.StatusBadRequest,
	errorhandling.ErrCodeUserNoUserFound:           http.StatusUnauthorized,
	errorhandling.ErrCodeUserAAL1Required:          http.StatusForbidden,
	errorhandling.ErrCodeUserAAL2Required:          http.StatusForbidden,
	errorhandling.ErrCodeTwilioInvalidOTP:          http.StatusBadRequest,
	errorhandling.ErrCodeWalletsNoWalletFound:      http.StatusNotFound,
	errorhandling.ErrCodeWalletsDuplicateWallet:    http.StatusConflict,
	errorhandling.ErrCodeWalletsWalletConflict:     http.StatusConflict,
	errorhandling.ErrCodeLinkedAccNotFound:         http.StatusNotFound,
	errorhandling.ErrCodeSignupDuplicatePhone:      http.StatusConflict,
	errorhandling.ErrCodeIdentitiesAlreadyExists:   http.StatusConflict,
	errorhandling.ErrCodePaymentsRequiredActions:   http.StatusBadRequest,
	errorhandling.ErrCodePaymentsInsufficientFunds: http.StatusBadRequest,
	errorhandling.ErrCodeKYCResubmissionRequired:   http.StatusForbidden,
}

func ToHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	log.Info("http error", zap.Error(err))

	appErr := errorhandling.ToAppError(err)

	status, ok := errorStatus[appErr.ErrorCode]
	if !ok {
		sentry.CaptureException(err)
		log.Error("unexpected error", zap.Error(err))
		WriteAppError(w, r, http.StatusInternalServerError, errorhandling.ErrCodeInternal, "Internal server error")
		return
	}

	if appErr.ErrorCode == errorhandling.ErrCodeInternal {
		sentry.CaptureException(err)
		log.Error("unexpected error", zap.Error(err))
	}

	WriteAppError(w, r, status, appErr.ErrorCode, http.StatusText(status))
}

func WriteAppError(w http.ResponseWriter, r *http.Request, status int, code errorhandling.AppErrorCode, message string) {
	reqID := appcontext.RequestIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(struct {
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
		ReqID     string `json:"req_id,omitempty"`
	}{code, message, reqID})
}
