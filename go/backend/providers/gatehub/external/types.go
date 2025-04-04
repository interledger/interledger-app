package external

import (
	"time"
)

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

	CardStatusActive           string = "Active"
	CardStatusBlocked          string = "Blocked"
	CardStatusTemporaryBlocked string = "TemporaryBlocked"
	CardStatusReplaced         string = "Replaced"
	CardStatusSoftDelete       string = "SoftDelete"
	CardStatusAccountBlocked   string = "AccountBlocked"
	CardStatusInCreation       string = "InCreation"
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

	Hub struct {
		Name                             string    `json:"name,omitempty"`
		CreatedAt                        time.Time `json:"created_at,omitempty"`
		UUID                             string    `json:"uuid,omitempty"`
		NameShort                        string    `json:"name_short,omitempty"`
		NameLegal                        string    `json:"name_legal,omitempty"`
		NameLegalShort                   string    `json:"name_legal_short,omitempty"`
		ContactEmail                     string    `json:"contact_email,omitempty"`
		Website                          string    `json:"website,omitempty"`
		Logo                             string    `json:"logo,omitempty"`
		Color                            string    `json:"color,omitempty"`
		State                            int       `json:"state,omitempty"`
		PhoneNumber                      string    `json:"phone_number,omitempty"`
		AddressStreet1Legal              string    `json:"address_street1_legal,omitempty"`
		AddressStreet2Legal              string    `json:"address_street2_legal,omitempty"`
		AddressCityLegal                 string    `json:"address_city_legal,omitempty"`
		AddressPostalCodeLegal           string    `json:"address_postal_code_legal,omitempty"`
		AddressCountryCodeLegal          string    `json:"address_country_code_legal,omitempty"`
		AddressSubdivisionLegal          string    `json:"address_subdivision_legal,omitempty"`
		AddressStreet1Correspondence     string    `json:"address_street1_correspondence,omitempty"`
		AddressStreet2Correspondence     string    `json:"address_street2_correspondence,omitempty"`
		AddressCityCorrespondence        string    `json:"address_city_correspondence,omitempty"`
		AddressPostalCodeCorrespondence  string    `json:"address_postal_code_correspondence,omitempty"`
		AddressCountryCodeCorrespondence string    `json:"address_country_code_correspondence,omitempty"`
		AddressSubdivisionCorrespondence string    `json:"address_subdivision_correspondence,omitempty"`
		Verified                         int       `json:"verified,omitempty"`
		ConnectedAt                      time.Time `json:"connected_at,omitempty"`
		VerifiedAt                       time.Time `json:"verified_at,omitempty"`
		Terms                            string    `json:"terms,omitempty"`
		IsShownInWallet                  bool      `json:"is_shown_in_wallet,omitempty"`
	}

	Verification struct {
		UUID         string `json:"uuid,omitempty"`
		State        int    `json:"state,omitempty"`
		Status       int    `json:"status,omitempty"`
		ProviderType string `json:"provider_type,omitempty"`
	}

	Document struct {
		ID                     int         `json:"id,omitempty"`
		CreatedAt              time.Time   `json:"created_at,omitempty"`
		UUID                   string      `json:"uuid,omitempty"`
		UpdatedAt              time.Time   `json:"updated_at,omitempty"`
		State                  int         `json:"state,omitempty"`
		Type                   string      `json:"type,omitempty"`
		Subtype                string      `json:"subtype,omitempty"`
		Value                  string      `json:"value,omitempty"`
		ProfileID              int         `json:"profile_id,omitempty"`
		Company                int         `json:"company,omitempty"`
		UserID                 int         `json:"user_id,omitempty"`
		ExpiryDate             time.Time   `json:"expiry_date,omitempty"`
		CompanyProfilesProfile interface{} `json:"company_profiles_profile,omitempty"`
	}

	Profile struct {
		UUID                   string      `json:"uuid,omitempty"`
		CreatedAt              time.Time   `json:"created_at,omitempty"`
		BirthDay               int         `json:"birth_day,omitempty"`
		BirthMonth             int         `json:"birth_month,omitempty"`
		BirthYear              int         `json:"birth_year,omitempty"`
		Gender                 string      `json:"gender,omitempty"`
		FirstName              string      `json:"first_name,omitempty"`
		MiddleName             string      `json:"middle_name,omitempty"`
		LastName               string      `json:"last_name,omitempty"`
		Citizenship            string      `json:"citizenship,omitempty"`
		AddressPostalCode      string      `json:"address_postal_code,omitempty"`
		AddressSubdivision     string      `json:"address_subdivision,omitempty"`
		AddressCountryCode     string      `json:"address_country_code,omitempty"`
		AddressCity            string      `json:"address_city,omitempty"`
		AddressStreet1         string      `json:"address_street1,omitempty"`
		AddressStreet2         string      `json:"address_street2,omitempty"`
		SSN                    interface{} `json:"ssn,omitempty"`
		Documents              []Document  `json:"documents,omitempty"`
		BirthCity              string      `json:"birth_city,omitempty"`
		BirthCountryCode       string      `json:"birth_country_code,omitempty"`
		TaxResidency           string      `json:"tax_residency,omitempty"`
		ExpectedVolume         string      `json:"expected_volume,omitempty"`
		OpeningReason          []string    `json:"opening_reason,omitempty"`
		SourceOfFunds          []string    `json:"source_of_funds,omitempty"`
		PEP                    bool        `json:"pep,omitempty"`
		EmploymentStatus       interface{} `json:"employment_status,omitempty"`
		TaxNumber              interface{} `json:"tax_number,omitempty"`
		NoTaxNumberReason      interface{} `json:"no_tax_number_reason,omitempty"`
		FATCA                  interface{} `json:"fatca,omitempty"`
		PaymentDestination1    interface{} `json:"payment_destination_1,omitempty"`
		PaymentDestination2    interface{} `json:"payment_destination_2,omitempty"`
		PaymentDestination3    interface{} `json:"payment_destination_3,omitempty"`
		PaymentsOriginCountry1 interface{} `json:"payments_origin_country_1,omitempty"`
		PaymentsOriginCountry2 interface{} `json:"payments_origin_country_2,omitempty"`
		PaymentsOriginCountry3 interface{} `json:"payments_origin_country_3,omitempty"`
		TempAddressStreet      interface{} `json:"temp_address_street,omitempty"`
		TempAddressStreet2     interface{} `json:"temp_address_street2,omitempty"`
		TempAddressCity        interface{} `json:"temp_address_city,omitempty"`
		TempAddressPostalCode  interface{} `json:"temp_address_postal_code,omitempty"`
		TempAddressCountry     interface{} `json:"temp_address_country,omitempty"`
		TempAddressState       interface{} `json:"temp_address_state,omitempty"`
	}

	User struct {
		UUID                  string         `json:"uuid,omitempty"`
		CreatedAt             time.Time      `json:"created_at,omitempty"`
		Photo                 interface{}    `json:"photo,omitempty"`
		PhoneNumber           interface{}    `json:"phone_number,omitempty"`
		UnverifiedPhoneNumber interface{}    `json:"unverified_phone_number,omitempty"`
		Email                 string         `json:"email,omitempty"`
		Deleted               int            `json:"deleted,omitempty"`
		UserType              int            `json:"user_type,omitempty"`
		MailSubs              bool           `json:"mail_subs,omitempty"`
		Hubs                  []Hub          `json:"hubs,omitempty"`
		Verifications         []Verification `json:"verifications,omitempty"`
		VerificationLevel     int            `json:"verification_level,omitempty"`
		Profile               Profile        `json:"profile,omitempty"`
		CompanyProfile        interface{}    `json:"companyProfile,omitempty"`
		Documents             []Document     `json:"documents,omitempty"`
		Href                  string         `json:"href,omitempty"`
	}

	Pagination struct {
		PageNumber uint `json:"pageNumber"`
		PageSize   uint `json:"pageSize"`
		TotalPages uint `json:"totalPages"`
	}

	Card struct {
		ID               string  `json:"id"`
		SourceID         string  `json:"sourceId"`
		AccountID        string  `json:"accountId"`
		AccountSourceID  string  `json:"accountSourceId"`
		CustomerID       string  `json:"customerId"`
		CustomerSourceID string  `json:"customerSourceId"`
		NameOnCard       string  `json:"nameOnCard"`
		ProductCode      string  `json:"productCode"`
		MaskedPan        string  `json:"maskedPan"`
		Status           string  `json:"status"`
		StatusReasonCode *string `json:"statusReasonCode"`
		LockLevel        *string `json:"lockLevel"`
		ExpiryDate       string  `json:"expiryDate"`
	}

	ListCardsResponse struct {
		Data       []Card     `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
)
