//go:build e2e

package main

const (
	mockXagoURL          = "http://localhost:24024"
	maxWaitSeconds       = 30
	defaultPolicy        = "5e2585a474b0e90012ce8ff1"
	defaultPubKey        = "test_public_key_12345"
	defaultSecret        = "test_secret_key_98765"
	defaultWebhookSecret = "test-webhook-secret"
	defaultWebhookAppID  = "xago-mock"
)

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

type createSubAccountResponse struct {
	AccountID          string                         `json:"accountId"`
	WalletID           string                         `json:"-"`
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
