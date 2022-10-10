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

type (
	// This will underpin a linked account and is used to create Machnet receive user accounts.
	ReceiveBankAccount struct {
		ID            string `db:"id"`
		WalletID      string `db:"wallet_id"`
		AccountNumber string `db:"account_number"`
		BankID        uint32 `db:"bank_id"`
		BranchID      uint32 `db:"branch_id"`
		CreatedAt     string `db:"created_at"`
		UpdatedAt     string `db:"updated_at"`
	}

	CreateReceiveBankAccountArgs struct {
		WalletID      string
		AccountNumber string
		BankID        uint32
		BranchID      uint32
	}
)

type (
	ReceiveUser struct {
		ID              string `db:"id"`
		SendUserID      string `db:"send_user_id"`
		ReceiveWalletID string `db:"receive_wallet_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}

	CreateReceiveUserArgs struct {
		ExternalID      string
		SendUserID      string
		ReceiveWalletID string
	}

	GetReceiveUserArgs struct {
		ReceiveWalletID string
		SendUserID      string
	}
)

type (
	ReceiveUserBankAccount struct {
		ID                   string `db:"id"`
		ReceiveUserID        string `db:"receive_user_id"`
		ReceiveBankAccountID string `db:"receive_bank_account_id"`
		CreatedAt            string `db:"created_at"`
		UpdatedAt            string `db:"updated_at"`
	}

	CreateReceiveUserBankAccountArgs struct {
		ExternalID           string
		ReceiveUserID        string
		ReceiveBankAccountID string
	}

	GetReceiveUserBankAccountArgs struct {
		ReceiveUserID        string
		ReceiveBankAccountID string
	}
)

type KYCStatus int

const (
	KYCStatusUnknown       KYCStatus = 0
	KYCStatusInProgress    KYCStatus = 1 // User's KYC is in progress.
	KYCStatusVerified      KYCStatus = 2 // User’s KYC is complete and user details are verified.
	KYCStatusRetry         KYCStatus = 3 // If user KYC is not complete and additional details have to be provided. Once the details have been collected, ‘Initiate KYC’ API needs to be called again.
	KYCStatusSuspended     KYCStatus = 4 // When a user's KYC is rejected during KYC process.
	KYCStatusReviewPending KYCStatus = 5 // When a user’s KYC process is in review state.
)

type Branch struct {
	ID   uint32
	Name string
}

type Bank struct {
	ID                        uint32
	Name                      string
	Branches                  []Branch
	Country                   string
	TransactionSupportedTypes []string
	ReceivingCurrency         []string
}
