package models

// LoginRequest represents the login request body
type LoginRequest struct {
	PolicyID string      `json:"policyId"`
	Fields   []FieldData `json:"fields"`
}

// FieldData represents a field in the login request
type FieldData struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	TokenValue string `json:"tokenValue"`
}

// CreateSubAccountRequest represents sub-account creation payload
type CreateSubAccountRequest struct {
	WalletID                  string `json:"walletId"`
	FirstName                 string `json:"firstName"`
	LastName                  string `json:"lastName"`
	Email                     string `json:"email"`
	MobileNumber              string `json:"mobileNumber"`
	IdentityType              string `json:"identityType"`
	IDNumber                  string `json:"idNumber"`
	PhysicalAddress           string `json:"physicalAddress"`
	ThirdPartyVerificationURL string `json:"thirdPartyVerificationUrl"`
}

// UpdateSubAccountRequest represents sub-account update payload
type UpdateSubAccountRequest struct {
	ThirdPartyVerificationURL string `json:"thirdPartyVerificationUrl"`
	IDNumber                  string `json:"idNumber"`
	PhysicalAddress           string `json:"physicalAddress"`
}

// BankDepositDetail represents a bank detail entry
type BankDepositDetail struct {
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	IBAN          string `json:"IBAN,omitempty"`
	SwiftBIC      string `json:"swiftBIC,omitempty"`
}

// BeneficiaryResponse represents beneficiary info in sub-account response
type BeneficiaryResponse struct {
	BeneficiaryID    string `json:"beneficiaryId"`
	BeneficiaryType  string `json:"beneficiaryType"`
	CurrencyID       string `json:"currencyId"`
	DepositReference string `json:"depositReference"`
	AccountNumber    string `json:"accountNumber"`
	BankName         string `json:"bankName"`
	AccountName      string `json:"accountName"`
}

// CreateSubAccountResponse represents creation response
type CreateSubAccountResponse struct {
	AccountID          string                         `json:"accountId"`
	DepositAddress     string                         `json:"depositAddress"`
	DepositTag         int                            `json:"depositTag"`
	BankDepositDetails map[string][]BankDepositDetail `json:"bankDepositDetails"`
	Beneficiaries      []BeneficiaryResponse          `json:"beneficiaries"`
}

// UpdateSubAccountResponse represents update response
type UpdateSubAccountResponse struct {
	AccountID string `json:"accountId"`
	Status    string `json:"status"`
}

// BalanceItem represents balance per currency
type BalanceItem struct {
	CurrencyCode string  `json:"currencyCode"`
	Available    float64 `json:"available"`
	Reserved     float64 `json:"reserved"`
	Total        float64 `json:"total"`
}

// BalanceResponse represents balance response for an account
type BalanceResponse struct {
	AccountID string        `json:"accountId"`
	Balances  []BalanceItem `json:"balances"`
}

// TestSetBalanceRequest represents a test-only balance set payload.
type TestSetBalanceRequest struct {
	AccountID    string  `json:"accountId"`
	WalletID     string  `json:"walletId"`
	CurrencyCode string  `json:"currencyCode"`
	Available    float64 `json:"available"`
	Reserved     float64 `json:"reserved"`
}

// TestBalanceDeltaRequest represents a test-only balance delta payload.
type TestBalanceDeltaRequest struct {
	AccountID    string  `json:"accountId"`
	WalletID     string  `json:"walletId"`
	CurrencyCode string  `json:"currencyCode"`
	Amount       float64 `json:"amount"`
}

// TestBalanceResponse represents a test-only balance response.
type TestBalanceResponse struct {
	Status string `json:"status"`
}

// CurrencyResponse represents available currency and bank details
type CurrencyResponse struct {
	CurrencyID    string `json:"currencyId"`
	CurrencyName  string `json:"currencyName"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	SwiftBIC      string `json:"swiftBIC"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
