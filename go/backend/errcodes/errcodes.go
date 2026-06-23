package errcodes

type AppErrorCode = string

const (
	ErrCodeInternal       AppErrorCode = "INTERNAL"
	ErrCodeUnauthorized   AppErrorCode = "UNAUTHORIZED"
	ErrCodeNotFound       AppErrorCode = "NOT_FOUND"
	ErrCodeConflict       AppErrorCode = "CONFLICT"
	ErrCodeForbidden      AppErrorCode = "FORBIDDEN"
	ErrCodeBadRequest     AppErrorCode = "BAD_REQUEST"
	ErrCodeGatewayTimeout AppErrorCode = "GATEWAY_TIMEOUT"

	ErrCodeValidation AppErrorCode = "VALIDATION"

	ErrCodeUserNoUserFound  AppErrorCode = "USER_NO_USER_FOUND"
	ErrCodeUserAAL1Required AppErrorCode = "USER_AAL1_REQUIRED"
	ErrCodeUserAAL2Required AppErrorCode = "USER_AAL2_REQUIRED"

	ErrCodeTwilioInvalidOTP AppErrorCode = "TWILIO_INVALID_OTP"

	ErrCodeWalletsNoWalletFound   AppErrorCode = "WALLETS_NO_WALLET_FOUND"
	ErrCodeWalletsDuplicateWallet AppErrorCode = "WALLETS_DUPLICATE_WALLET"
	ErrCodeWalletsWalletConflict  AppErrorCode = "WALLETS_WALLET_CONFLICT"

	ErrCodeLinkedAccNotFound AppErrorCode = "LINKEDACC_NOT_FOUND"

	ErrCodeSignupDuplicatePhone AppErrorCode = "SIGNUP_DUPLICATE_PHONE"

	ErrCodeIdentitiesAlreadyExists AppErrorCode = "IDENTITIES_ALREADY_EXISTS"

	ErrCodePaymentsRequiredActions   AppErrorCode = "PAYMENTS_REQUIRED_ACTIONS"
	ErrCodePaymentsInsufficientFunds AppErrorCode = "PAYMENTS_INSUFFICIENT_FUNDS"

	ErrCodeKYCResubmissionRequired AppErrorCode = "KYC_RESUBMISSION_REQUIRED"

	ErrCodeUserInvalidTotpCode   AppErrorCode = "USER_INVALID_TOTP_CODE"
	ErrCodeUserTotpNotConfigured AppErrorCode = "USER_TOTP_NOT_CONFIGURED"
	ErrCodeUserInvalidTotpConfig AppErrorCode = "USER_INVALID_TOTP_CONFIG"

	ErrCodeEmailSupportInboxNotConfigured AppErrorCode = "EMAIL_SUPPORT_INBOX_NOT_CONFIGURED"

	ErrCodeAccountDeletionAlreadyRequested AppErrorCode = "ACCOUNT_DELETION_ALREADY_REQUESTED"
)
