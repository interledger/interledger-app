package external

const (
	StatusVerified   = "VERIFIED"
	StatusUnverified = "UNVERIFIED"
	StatusFailed     = "FAILED"
	StatusPending    = "PENDING"

	SendUser    = "SEND"
	ReceiveUser = "RECEIVE"

	TypeCard = "CARD"
	TypeBank = "Bank"

	DeliveryStatusNone      = "NONE"
	DeliveryStatusRequested = "DELIVERY_REQUESTED"
)

type User struct {
	ID                string             `json:"id,omitempty"`
	FirstName         string             `json:"first_name,omitempty"`
	MiddleName        string             `json:"middle_name,omitempty"`
	LastName          string             `json:"last_name,omitempty"`
	Email             string             `json:"email,omitempty"`
	Gender            string             `json:"gender,omitempty"`
	DateOfBirth       string             `json:"date_of_birth,omitempty"`
	AddressLine1      string             `json:"address_line1,omitempty"`
	AddressLine2      string             `json:"address_line2,omitempty"`
	MobilePhone       string             `json:"mobile_phone,omitempty"`
	City              string             `json:"city,omitempty"`
	Zipcode           string             `json:"zipcode,omitempty"`
	State             string             `json:"state,omitempty"`
	Country           string             `json:"country,omitempty"`
	IPAddress         string             `json:"ip_address,omitempty"`
	Type              string             `json:"type,omitempty"`
	PhysicalDocuments []PhysicalDocument `json:"physical_documents,omitempty"`
	Status            string             `json:"status,omitempty"`
	SendUserID        string             `json:"send_user_id,omitempty"`
	Business          bool               `json:"business,omitempty"`
	BusinessType      string             `json:"business_type,omitempty"`
}

type PhysicalDocument struct {
	DocumentValue     string `json:"document_value,omitempty"`
	DocumentValueBack string `json:"document_value_back,omitempty"`
	DocumentType      string `json:"document_type,omitempty"`
	DocumentCountry   string `json:"country,omitempty"`
	DocumentState     string `json:"state,omitempty"`
}

type CipInfo struct {
	Gender       string `json:"GENDER,omitempty"`
	Country      string `json:"COUNTRY,omitempty"`
	State        string `json:"STATE,omitempty"`
	LastName     string `json:"LAST_NAME,omitempty"`
	Occupation   string `json:"OCCUPATION,omitempty"`
	City         string `json:"CITY,omitempty"`
	ZipCode      string `json:"ZIP_CODE,omitempty"`
	DateOfBirth  string `json:"DATE_OF_BIRTH,omitempty"`
	IDDoc        string `json:"ID_DOC,omitempty"`
	PhoneNumber  string `json:"PHONE_NUMBER,omitempty"`
	Email        string `json:"EMAIL,omitempty"`
	AddressLine1 string `json:"ADDRESS_LINE1,omitempty"`
	AddressLine2 string `json:"ADDRESS_LINE2,omitempty"`
	MiddleName   string `json:"MIDDLE_NAME,omitempty"`
	FirstName    string `json:"FIRST_NAME,omitempty"`
	IPAddress    string `json:"IP_ADDRESS,omitempty"`
}

type VerificationStatus struct {
	CipInfo   CipInfo `json:"cip_info,omitempty"`
	KycStatus string  `json:"kyc_status,omitempty"`
	UserID    string  `json:"user_id,omitempty"`
}

type WidgetTokenResponse struct {
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

type FundingSource struct {
	ID                 string `json:"id,omitempty"`
	UserID             string `json:"user_id,omitempty"`
	FundingsourceName  string `json:"funding_source_name,omitempty"`
	FundingsourceType  string `json:"funding_source_type,omitempty"`
	AccountNumber      string `json:"account_number,omitempty"`
	InstitutionName    string `json:"institution_name,omitempty"`
	VerificationStatus string `json:"verification_status,omitempty"`
}

type Transaction struct {
	ID                string  `json:"id,omitempty"`
	UserID            string  `json:"user_id,omitempty"`
	FromAmount        float64 `json:"from_amount,omitempty"`
	FromCurrency      string  `json:"from_currency,omitempty"`
	ToAmount          float64 `json:"to_amount,omitempty"`
	ToCurrency        string  `json:"to_currency,omitempty"`
	FeeAmount         float64 `json:"fee_amount,omitempty"`
	ExchangeRate      float64 `json:"exchange_rate,omitempty"`
	FromFundID        string  `json:"from_fund_id,omitempty"`
	FundingsourceType string  `json:"funding_source_type,omitempty"`
	DeliveryStatus    string  `json:"delivery_status,omitempty"`
	// TODO: to field
}

type DeliveryRequest struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
}
