package external

import "time"

type Product string

var (
	OnOffRamp  Product = "OnOffRamp"
	Onboarding Product = "Onboarding"
	Exchange   Product = "Exchange"

	TransactionTypeWithdrawal int = 0
	TransactionTypeDeposit    int = 1
	TransactionTypeHosted     int = 2
	TransactionTypeExchange   int = 3

	TransactionStatusPending        int = 0
	TransactionStatusProcessing     int = 1
	TransactionStatusUnmatched      int = 3
	TransactionStatusReturning      int = 5
	TransactionStatusManualReview   int = 10
	TransactionStatusCompleted      int = 100
	TransactionStatusFailed         int = 101
	TransactionStatusUserCancelled  int = 102
	TransactionStatusAdminCancelled int = 103
)

type (
	IssueTokenReqeust struct {
		Scope []string `json:"scope"`
	}

	IssueTokenResponse struct {
		Token     string `json:"token,omitempty"`
		ExpiresAt string `json:"expires,omitempty"`
	}

	CreateUserRequest struct {
		Email string `json:"email,omitempty"`
	}

	CreateUserResponse struct {
		ID                 string                 `json:"id"`
		CreatedAt          time.Time              `json:"createdAt"`
		UpdatedAt          time.Time              `json:"updatedAt"`
		ActivatedAt        time.Time              `json:"activatedAt"`
		Email              string                 `json:"email"`
		Secret2FA          bool                   `json:"secret2fa"`
		Type2FA            string                 `json:"type2fa"`
		Activated          bool                   `json:"activated"`
		Role               string                 `json:"role"`
		Meta               map[string]interface{} `json:"meta"`
		LastPasswordChange time.Time              `json:"lastPasswordChange"`
		Features           []string               `json:"features"`
		Managed            bool                   `json:"managed"`
		ManagedBy          string                 `json:"managedBy"`
	}

	Wallet struct {
		UUID               string                 `json:"uuid"`
		CreatedAt          time.Time              `json:"created_at"`
		UpdatedAt          *time.Time             `json:"updated_at"` // Using pointer to handle null
		SynchronizedAt     *time.Time             `json:"synchronized_at"`
		ExpiresAt          *time.Time             `json:"expires_at"`
		Address            string                 `json:"address"`
		RegularAddress     string                 `json:"regular_address"`
		MigrationStatus    int                    `json:"migration_status"`
		MigrationDate      *time.Time             `json:"migration_date"`
		MultisignAddress   string                 `json:"multisign_address"`
		MultisignStatus    string                 `json:"multisign_status"`
		Primary            bool                   `json:"primary"`
		Type               int                    `json:"type"`
		Deleted            bool                   `json:"deleted"`
		Name               string                 `json:"name"`
		Version            int                    `json:"version"`
		Enabled            bool                   `json:"enabled"`
		Active             bool                   `json:"active"`
		Meta               map[string]interface{} `json:"meta"` // Using a map to handle an empty JSON object
		ThirdPartyProvider *string                `json:"third_party_provider"`
	}

	GetUserWalletsResponse struct {
		UUID      string  `json:"uuid"`
		Email     string  `json:"email"`
		Paystring *string `json:"paystring"` // Using pointer to handle null
		Enabled   int     `json:"enabled"`   // Assuming an int based on the example, could also be a bool depending on the actual data definition
		// PrimaryVault *string   `json:"primary_vault"`
		// DisplayVault *string   `json:"display_vault"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Wallets   []Wallet  `json:"wallets"`
	}

	Vault struct {
		UUID      string `json:"uuid"`
		Name      string `json:"name"`
		AssetCode string `json:"asset_code"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	WalletBalance struct {
		Available string `json:"available"`
		Pending   string `json:"pending"`
		Total     string `json:"total"`
		Vault     Vault  `json:"vault"`
	}

	CreateTransactionRequest struct {
		SendingUserID    string  `json:"-"`
		Amount           float64 `json:"amount"`
		SendingAddress   string  `json:"sending_address"`
		ReceivingAddress string  `json:"receiving_address"`
		Message          string  `json:"message"`
		Type             int     `json:"type"`
		VaultID          string  `json:"vault_uuid"`
	}

	Transaction struct {
		ID              string `json:"uuid"`
		CreatedAt       string `json:"created_at"`
		CompletedAt     string `json:"completed_at"`
		Amount          string `json:"amount"`
		Total           string `json:"total_amount"`
		Fee             string `json:"fee"`
		SendingWallet   Wallet `json:"sending_wallet"`
		ReceivingWallet Wallet `json:"receiving_wallet"`
		Vault           Vault  `json:"vault"`
		Status          int    `json:"status"`
		SubStatus       int    `json:"substatus"`
		Type            int    `json:"type"`
	}
)
