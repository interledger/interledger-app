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
	AccountID     string  `json:"accountId"`
	WalletID      string  `json:"walletId"`
	CurrencyCode  string  `json:"currencyCode"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transactionId"` // Optional: references pre-created transaction for backend verification
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

// DepositFields contains bank account deposit details
type DepositFields struct {
	BankName       string `json:"bankName"`
	AccountName    string `json:"accountName"`
	AccountNumber  string `json:"accountNumber"`
	BankAddress    string `json:"bankAddress,omitempty"`
	AccountAddress string `json:"accountAddress,omitempty"`
	BranchCode     string `json:"branchCode"`
	SwiftBIC       string `json:"swiftBIC,omitempty"`
}

// BankingProvider represents a banking provider with deposit availability
type BankingProvider struct {
	Name             string        `json:"name"`
	DepositAvailable bool          `json:"depositAvailable"`
	DepositFields    DepositFields `json:"depositFields"`
}

// CurrencyNested represents currency with nested banking providers
// This format matches what the backend expects (see go/backend/providers/xago/external/types.go)
type CurrencyNested struct {
	CurrencyCode     string            `json:"currencyCode"`
	Name             string            `json:"name,omitempty"`
	Symbol           string            `json:"symbol,omitempty"`
	DepositEnabled   bool              `json:"depositEnabled"`
	WithdrawEnabled  bool              `json:"withdrawEnabled"`
	MarketEnabled    bool              `json:"marketEnabled,omitempty"`
	BankingProviders []BankingProvider `json:"bankingProviders"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// AddBeneficiaryRequest represents the payload for adding a beneficiary
type AddBeneficiaryRequest struct {
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	CurrencyCode  string `json:"currencyCode"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	Reference     string `json:"reference"`
	IsOwn         bool   `json:"isOwn"`
}

// BeneficiaryItem represents a single beneficiary in API responses
type BeneficiaryItem struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	CurrencyCode  string `json:"currencyCode"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	Reference     string `json:"reference"`
	IsOwn         bool   `json:"isOwn"`
	Status        string `json:"status"`
}

// BeneficiaryPagination represents pagination info in list responses
type BeneficiaryPagination struct {
	Limit         int `json:"limit"`
	Page          int `json:"page"`
	NumberOfPages int `json:"numberOfPages"`
	Total         int `json:"total"`
}

// ListBeneficiariesResponse represents the paginated list of beneficiaries
type ListBeneficiariesResponse struct {
	Data       []BeneficiaryItem     `json:"data"`
	Pagination BeneficiaryPagination `json:"pagination"`
}

// CreateTransferRequest represents the payload for creating a transfer
type CreateTransferRequest struct {
	Amount         float64 `json:"amount"`
	CurrencyCode   string  `json:"currencyCode"`
	BeneficiaryID  string  `json:"beneficiaryId"`
	Reference      string  `json:"reference,omitempty"`
	IdempotencyKey string  `json:"idempotencyKey,omitempty"`
}

// CreateTransferResponse represents the response when creating a transfer
type CreateTransferResponse struct {
	TransactionID string `json:"transactionId"`
}

// TransactionItem represents a single transaction in list responses
type TransactionItem struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	CurrencyCode  string  `json:"currencyCode"`
	BeneficiaryID string  `json:"beneficiaryId"`
	Reference     string  `json:"reference"`
	CreatedAt     string  `json:"createdAt"`
	SettledAt     string  `json:"settledAt,omitempty"`
}

// TransactionPagination represents pagination info for transaction lists
type TransactionPagination struct {
	Limit         int `json:"limit"`
	Page          int `json:"page"`
	NumberOfPages int `json:"numberOfPages"`
	Total         int `json:"total"`
}

// ListTransactionsResponse represents the paginated list of transactions
type ListTransactionsResponse struct {
	Data       []TransactionItem     `json:"data"`
	Pagination TransactionPagination `json:"pagination"`
}

// GetTransactionResponse represents details of a single transaction
type GetTransactionResponse struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	CurrencyCode  string  `json:"currencyCode"`
	BeneficiaryID string  `json:"beneficiaryId"`
	Reference     string  `json:"reference"`
	CreatedAt     string  `json:"createdAt"`
	SettledAt     string  `json:"settledAt,omitempty"`
}

// TestDepositRequest represents the payload for simulating a test deposit
type TestDepositRequest struct {
	AccountID        string  `json:"accountId"`
	Amount           float64 `json:"amount"`
	CurrencyCode     string  `json:"currencyCode"`
	DepositReference string  `json:"depositReference,omitempty"`
}

// TestDepositResponse represents the response from simulating a test deposit
type TestDepositResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
}

// DepositItem represents a single deposit in list responses
type DepositItem struct {
	TransactionID        string  `json:"transactionId"`
	AccountID            string  `json:"accountId"`
	Amount               float64 `json:"amount"`
	CurrencyCode         string  `json:"currencyCode"`
	DepositReference     string  `json:"depositReference,omitempty"`
	TransactionReference string  `json:"transactionReference,omitempty"`
	Status               string  `json:"status"`
	Code                 int     `json:"code"`
	CreatedAt            string  `json:"createdAt"`
	SettledAt            string  `json:"settledAt,omitempty"`
}

// DepositPagination represents pagination info for deposit lists
type DepositPagination struct {
	Limit         int `json:"limit"`
	Page          int `json:"page"`
	NumberOfPages int `json:"numberOfPages"`
	Total         int `json:"total"`
}

// ListDepositsResponse represents the paginated list of deposits
type ListDepositsResponse struct {
	Data       []DepositItem     `json:"data"`
	Pagination DepositPagination `json:"pagination"`
}

// ConvertCurrencyPairEnum identifies a currency conversion direction.
type ConvertCurrencyPairEnum string

const (
	ZARtoEUR ConvertCurrencyPairEnum = "ZAR/EUR"
	EURtoZAR ConvertCurrencyPairEnum = "EUR/ZAR"
)

// ConvertCurrencyRequest is the payload for both estimate and actual conversion.
type ConvertCurrencyRequest struct {
	ConvertCurrencyPair ConvertCurrencyPairEnum `json:"convertCurrencyPair"`

	Amount              float64 `json:"amount"`
	EstimateCalculation bool    `json:"estimateCalculation"`
}

// EstimateConvertCurrencyResponse is the response for an estimate request.
type EstimateConvertCurrencyResponse struct {
	BuyAveragePrice float64 `json:"buyAveragePrice"`
	BuyOrders       float64 `json:"buyOrders"`
	EstimatedRate   float64 `json:"estimatedRate"`
	FinalBuyAmount  float64 `json:"finalBuyAmount"`
	FinalSellAmount float64 `json:"finalSellAmount"`
	QuoteAmount     float64 `json:"quoteAmount"`
	ReceivedAmount  float64 `json:"receivedAmount"`
	SellOrders      float64 `json:"sellOrders"`
}

// DepositWebhookPayload represents the webhook payload sent when a deposit completes
type DepositWebhookPayload struct {
	AccountID            string  `json:"accountId"`
	Amount               float64 `json:"amount"`
	CurrencyCode         string  `json:"currencyCode"`
	TransactionID        string  `json:"transactionId"`
	TransactionReference string  `json:"transactionReference,omitempty"`
	Status               string  `json:"status"`
	Code                 int     `json:"code"`
	CreatedAt            string  `json:"createdAt"`
	SettledAt            string  `json:"settledAt,omitempty"`
}
