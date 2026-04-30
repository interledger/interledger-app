package errorhandling

// shared error handling

type AppErrorCode = string

const (
	ErrCodeInternal       AppErrorCode = "INTERNAL"
	ErrCodeUnauthorized   AppErrorCode = "UNAUTHORIZED"
	ErrCodeNotFound       AppErrorCode = "NOT_FOUND"
	ErrCodeConflict       AppErrorCode = "CONFLICT"
	ErrCodeForbidden      AppErrorCode = "FORBIDDEN"
	ErrCodeBadRequest     AppErrorCode = "BAD_REQUEST"
	ErrCodeGatewayTimeout AppErrorCode = "GATEWAY_TIMEOUT"

	ErrCodeUserNoUserFound AppErrorCode = "USER_NO_USER_FOUND"

	ErrCodeWalletsNoWalletFound   AppErrorCode = "WALLETS_NO_WALLET_FOUND"
	ErrCodeWalletsDuplicateWallet AppErrorCode = "WALLETS_DUPLICATE_WALLET"
	ErrCodeWalletsWalletConflict  AppErrorCode = "WALLETS_WALLET_CONFLICT"

	ErrCodeLinkedAccNotFound AppErrorCode = "LINKEDACC_NOT_FOUND"

	ErrCodeKYCResubmissionRequired AppErrorCode = "KYC_RESUBMISSION_REQUIRED"
)
