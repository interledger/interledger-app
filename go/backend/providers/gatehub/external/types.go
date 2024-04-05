package external

import "time"

type Product string

var (
	OnOffRamp  Product = "OnOffRamp"
	Onboarding Product = "Onboarding"
	Exchange   Product = "Exchange"
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
		UUID         string    `json:"uuid"`
		Email        string    `json:"email"`
		Paystring    *string   `json:"paystring"` // Using pointer to handle null
		Enabled      int       `json:"enabled"`   // Assuming an int based on the example, could also be a bool depending on the actual data definition
		PrimaryVault *string   `json:"primary_vault"`
		DisplayVault *string   `json:"display_vault"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Wallets      []Wallet  `json:"wallets"`
	}
)
