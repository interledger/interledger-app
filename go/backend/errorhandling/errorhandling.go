package errorhandling

// shared error handling

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
