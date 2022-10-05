package machnet

const ProviderName = "machnet"

type User struct {
	ID        string    `db:"id"`
	WalletID  string    `db:"wallet_id"`
	CreatedAt string    `db:"created_at"`
	UpdatedAt string    `db:"updated_at"`
	KYCStatus KYCStatus `db:"kyc_status"`
}

type CreateArgs struct {
	WalletID   string
	ExternalID string
}

type WidgetToken struct {
	Value            string
	ExpiresInMinutes int
	UserID           string
}

type KYCStatus int

const (
	KYCStatusUnknown       KYCStatus = 0
	KYCStatusInProgress    KYCStatus = 1 // User's KYC is in progress.
	KYCStatusVerified      KYCStatus = 2 // User’s KYC is complete and user details are verified.
	KYCStatusRetry         KYCStatus = 3 // If user KYC is not complete and additional details have to be provided. Once the details have been collected, ‘Initiate KYC’ API needs to be called again.
	KYCStatusSuspended     KYCStatus = 4 // When a user's KYC is rejected during KYC process.
	KYCStatusReviewPending KYCStatus = 5 // When a user’s KYC process is in review state.
)
