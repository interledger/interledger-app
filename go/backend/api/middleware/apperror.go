package middleware

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeInternal     = "INTERNAL"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeGatewayTimeout = "GATEWAY_TIMEOUT"

	ErrCodeUserNoUserFound  = "USER_NO_USER_FOUND"

	ErrCodeWalletsNoWalletFound   = "WALLETS_NO_WALLET_FOUND"
	ErrCodeWalletsDuplicateWallet = "WALLETS_DUPLICATE_WALLET"
	ErrCodeWalletsWalletConflict  = "WALLETS_WALLET_CONFLICT"

	ErrCodeLinkedAccNotFound = "LINKEDACC_NOT_FOUND"

	ErrCodeKYCResubmissionRequired = "KYC_RESUBMISSION_REQUIRED"

	ErrCodeGatehubNotFound = "GATEHUB_NOT_FOUND"
	ErrCodeGatehubInternal = "INTERNAL"
	ErrCodeGatehubTimedOut = "GATEHUB_TIMED_OUT"
)

type appError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	ReqID     string `json:"req_id,omitempty"`
}

func WriteAppError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	reqID := RequestIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(appError{
		ErrorCode: code,
		Message:   message,
		ReqID:     reqID,
	})
}
