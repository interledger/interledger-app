package external

type CreateIntentReq struct {
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	City           string `json:"city"`
	State          string `json:"state"`         // Two letter abbreviation.
	PostalCode     string `json:"postal_code"`   // 5  digit string
	DateOfBirth    string `json:"date_of_birth"` // YYYY-MM-DD
	SocialSecurity string `json:"ssn"`           // No dashes
	IPAddress      string `json:"ip_address"`
}

type CreateIntentResp struct {
	ID string `json:"id"`
}

type Intent struct {
	ID                 string `json:"id"`
	UserID             string `json:"user_id"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	PreferredFirstName string `json:"preferred_first_name"`
	PreferredLastName  string `json:"preferred_last_name"`
	PreferredPronouns  string `json:"preferred_pronouns"`
	Status             string `json:"status"`
	KycType            string `json:"kyc_type"`
}

type GetVerificationTokenReq struct {
	Provider     string `json:"provider"`
	ProviderData struct {
		CustomerID string `json:"customer_id"`
	} `json:"provider_data"`
	ClientID     string `json:"client_id"`
	UserIntentID string `json:"user_intent_id"`
}

type VerificationTokenResp struct {
	Token string `json:"token"`
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

type CreateCardArgs struct {
	CardNumber       string `json:"card_number"`
	CardSecurityCode string `json:"card_security_code"`
	ExpirationDate   string `json:"expiration_date"` // Format MM/YY
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	StreetLine1      string `json:"street_line_1"`
	StreetLine2      string `json:"street_line_2"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZipCode          string `json:"zip_code"`
	AddedByUser      bool   `json:"added_by_user"`
}

type UserCard struct {
	ID              string `json:"id"`
	AddressVerified bool   `json:"address_verified"`
	CardCompany     string `json:"card_company"`
	City            string `json:"city"`
	Created         string `json:"created"`
	ExpirationDate  string `json:"expiration_date"`
	FirstName       string `json:"first_name"`
	FirstSixDigits  string `json:"first_six_digits"`
	LastFourDigits  string `json:"last_four_digits"`
	LastName        string `json:"last_name"`
	CardType        string `json:"card_type"`
	PullEnabled     bool   `json:"pull_enabled"`
	PushEnabled     bool   `json:"push_enabled"`
	Removed         bool   `json:"removed"`
	ReviewStatus    string `json:"review_status"`
	State           string `json:"state"`
	Status          string `json:"status"`
	StreetLine1     string `json:"street_line_1"`
	StreetLine2     string `json:"street_line_2"`
	ZipCode         string `json:"zip_code"`
}

type CardBin struct {
	Bin                 string `json:"bin"`
	CardBrand           string `json:"card_brand"`
	CardType            string `json:"card_type"`
	InterchangeCategory string `json:"interchange_category"`
	IssuerName          string `json:"issuer_name"`
	SettlementNetwork   string `json:"settlement_network"`
}

type CreateAccountArgs struct {
	InstitutionID   string      `json:"institution_id"`
	BankAccountType AccountType `json:"bank_account_type"`
	Name            string      `json:"name"`
	AccountNumber   string      `json:"account_number"`
	RoutingNumber   string      `json:"routing_number"`
}

type AccountType string

const (
	AccountTypeSavings  = "savings"
	AccountTypeChecking = "checking"
)

type UserAccount struct {
	ID               string      `json:"id"`
	OfficialName     string      `json:"official_name"`
	Name             string      `json:"name"`
	Mask             string      `json:"mask"`
	InstitutionName  string      `json:"institution_name"`
	InstitutionLogo  string      `json:"institution_logo"`
	Type             AccountType `json:"type"`
	Subtype          string      `json:"subtype"`
	ConnectionStatus string      `json:"connection_status"`
}

type CardToAccountArgs struct {
	IdempotencyKey      string      `json:"-"`
	Name                string      `json:"name"`
	Amount              float64     `json:"amount"`
	ClientCorrelationID string      `json:"client_correlation_id" validate:"len=8"` // Exactly 8 Chars
	DebitFeePercent     int         `json:"debit_fee_percent"`
	Card                Source      `json:"source"`
	Account             Destination `json:"destination"`
}

type AccountToCardArgs struct {
	IdempotencyKey  string      `json:"-"`
	Transfer        Transfer    `json:"transfer"`
	Account         Source      `json:"source"`
	Card            Destination `json:"destination"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
	Amount          float64     `json:"amount"`
	Name            string      `json:"name"`
}

type Transfer struct {
	Type    string `json:"type"` // ach_credit |
	Addenda string `json:"addenda"`
}

type Source struct {
	ID string `json:"id"`
}

type Destination struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type AccountToCardResp struct {
	ID              string      `json:"id"`
	Status          string      `json:"status"`
	Name            string      `json:"name"`
	AmountUSD       float64     `json:"amount"`
	Source          Source      `json:"source"`
	Destination     Destination `json:"destination"`
	StartDate       string      `json:"start_date"`
	Created         string      `json:"created"`
	Active          bool        `json:"active"`
	PaymentRoute    string      `json:"payment_route"`
	Type            string      `json:"type"`
	Blocked         bool        `json:"blocked"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
}

type CardToAccountResp struct {
	ID              string      `json:"id"`
	Status          string      `json:"status"`
	Name            string      `json:"name"`
	AmountUSD       float64     `json:"amount"`
	Source          Source      `json:"source"`
	Destination     Destination `json:"destination"`
	StartDate       string      `json:"start_date"`
	Created         string      `json:"created"`
	Active          bool        `json:"active"`
	PaymentRoute    string      `json:"payment_route"`
	Type            string      `json:"type"`
	Blocked         bool        `json:"blocked"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
}

type Transaction struct {
	ID                    string  `json:"id"`
	RoutineType           string  `json:"routine_type"`
	RoutineName           string  `json:"routine_name"`
	RoutineID             string  `json:"routine_id"`
	ClientCorrelationID   string  `json:"client_correlation_id"`
	SourceID              string  `json:"source_id"`
	DestinationID         string  `json:"destination_id"`
	DestinationUserID     string  `json:"destination_user_id"`
	Amount                float64 `json:"amount"`
	PaymentType           string  `json:"payment_type"`
	Initiated             string  `json:"initiated"`
	Updated               string  `json:"updated"`
	EstimatedClearingDate string  `json:"estimated_clearing_date"`
	AstraSettlementReason string  `json:"astra_settlement_reason"`
	FailureReason         string  `json:"failure_reason"`
	Chargeback            struct {
		ActionStatus           string  `json:"action_status"`
		CbID                   string  `json:"cb_id"`
		Created                string  `json:"created"`
		ExceptionCode          string  `json:"exception_code"`
		ExceptionDate          string  `json:"exception_date"`
		ExceptionDescription   string  `json:"exception_description"`
		ExceptionID            string  `json:"exception_id"`
		ExceptionSettledAmount float64 `json:"exception_settled_amount"`
		ExceptionType          string  `json:"exception_type"`
		MerchantReferenceID    string  `json:"merchant_reference_id"`
		NetworkID              string  `json:"network_id"`
		OriginalCreationDate   string  `json:"original_creation_date"`
		OriginalProcessedDate  string  `json:"original_processed_date"`
		OriginalSettledAmount  float64 `json:"original_settled_amount"`
		StatusDate             string  `json:"status_date"`
		Updated                string  `json:"updated"`
		UserID                 string  `json:"user_id"`
	} `json:"chargeback"`
	Status string `json:"status"`
}
