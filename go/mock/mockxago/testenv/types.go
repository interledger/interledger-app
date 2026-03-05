package main

import (
	"net/http"
	"time"
)

const (
	mockXagoURL          = "http://localhost:24024"
	maxWaitSeconds       = 30
	defaultPolicy        = "5e2585a474b0e90012ce8ff1"
	defaultPubKey        = "test_public_key_12345"
	defaultSecret        = "test_secret_key_98765"
	defaultWebhookSecret = "test-webhook-secret"
	defaultWebhookAppID  = "xago-mock"
)

type createSubAccountResponse struct {
	AccountID          string                         `json:"accountId"`
	WalletID           string                         `json:"-"` // Not in API, but needed for tests
	DepositAddress     string                         `json:"depositAddress"`
	DepositTag         int                            `json:"depositTag"`
	BankDepositDetails map[string][]bankDepositDetail `json:"bankDepositDetails"`
	Beneficiaries      []beneficiaryResponse          `json:"beneficiaries"`
}

type bankDepositDetail struct {
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	SwiftBIC      string `json:"swiftBIC"`
}

type beneficiaryResponse struct {
	BeneficiaryID    string `json:"beneficiaryId"`
	BeneficiaryType  string `json:"beneficiaryType"`
	CurrencyID       string `json:"currencyId"`
	DepositReference string `json:"depositReference"`
	AccountNumber    string `json:"accountNumber"`
	BankName         string `json:"bankName"`
	AccountName      string `json:"accountName"`
}

type balanceResponse struct {
	AccountID string        `json:"accountId"`
	Balances  []balanceItem `json:"balances"`
}

type balanceItem struct {
	CurrencyCode string  `json:"currencyCode"`
	Available    float64 `json:"available"`
	Reserved     float64 `json:"reserved"`
	Total        float64 `json:"total"`
}

type currencyResponse struct {
	CurrencyID    string `json:"currencyId"`
	CurrencyName  string `json:"currencyName"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
	SwiftBIC      string `json:"swiftBIC"`
}

// Nested currency response types (for backend compatibility)
type depositFields struct {
	BankName       string `json:"bankName"`
	AccountName    string `json:"accountName"`
	AccountNumber  string `json:"accountNumber"`
	BankAddress    string `json:"bankAddress,omitempty"`
	AccountAddress string `json:"accountAddress,omitempty"`
	BranchCode     string `json:"branchCode"`
	SwiftBIC       string `json:"swiftBIC,omitempty"`
}

type bankingProvider struct {
	Name             string        `json:"name"`
	DepositAvailable bool          `json:"depositAvailable"`
	DepositFields    depositFields `json:"depositFields"`
}

type currencyNested struct {
	CurrencyCode     string            `json:"currencyCode"`
	Name             string            `json:"name,omitempty"`
	Symbol           string            `json:"symbol,omitempty"`
	DepositEnabled   bool              `json:"depositEnabled"`
	WithdrawEnabled  bool              `json:"withdrawEnabled"`
	MarketEnabled    bool              `json:"marketEnabled,omitempty"`
	BankingProviders []bankingProvider `json:"bankingProviders"`
}

// Helper to convert nested to flat format for backward compatibility
func (c *currencyNested) toFlat() *currencyResponse {
	if len(c.BankingProviders) == 0 {
		return nil
	}
	provider := c.BankingProviders[0]
	return &currencyResponse{
		CurrencyID:    c.CurrencyCode,
		CurrencyName:  c.Name,
		BankName:      provider.DepositFields.BankName,
		AccountName:   provider.DepositFields.AccountName,
		AccountNumber: provider.DepositFields.AccountNumber,
		BranchCode:    provider.DepositFields.BranchCode,
		SwiftBIC:      provider.DepositFields.SwiftBIC,
	}
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type addBeneficiaryResponse struct {
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

type beneficiaryPagination struct {
	Limit         int `json:"limit"`
	Page          int `json:"page"`
	NumberOfPages int `json:"numberOfPages"`
	Total         int `json:"total"`
}

type listBeneficiariesResponse struct {
	Data       []addBeneficiaryResponse `json:"data"`
	Pagination beneficiaryPagination    `json:"pagination"`
}

type depositResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
}

type companyDepositItem struct {
	TransactionID string  `json:"transactionId"`
	AccountID     string  `json:"accountId"`
	Amount        float64 `json:"amount"`
	CurrencyCode  string  `json:"currencyCode"`
	Status        string  `json:"status"`
	Code          int     `json:"code"`
	CreatedAt     string  `json:"createdAt"`
	SettledAt     string  `json:"settledAt"`
}

type companyDepositPagination struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
	Total int `json:"total"`
}

type listCompanyDepositsResponse struct {
	Data       []companyDepositItem     `json:"data"`
	Pagination companyDepositPagination `json:"pagination"`
}

type webhookPayload struct {
	AccountID            string  `json:"accountId"`
	Amount               float64 `json:"amount"`
	CurrencyCode         string  `json:"currencyCode"`
	TransactionID        string  `json:"transactionId"`
	TransactionReference string  `json:"transactionReference"`
	Status               string  `json:"status"`
	Code                 int     `json:"code"`
	CreatedAt            string  `json:"createdAt"`
	SettledAt            string  `json:"settledAt"`
}

type webhookEvent struct {
	Body       webhookPayload
	Headers    http.Header
	RawBody    []byte
	ReceivedAt time.Time
}
