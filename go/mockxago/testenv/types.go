package main

import (
	"net/http"
)

const (
	baseURL        = "http://localhost:28080"
	maxWaitSeconds = 30
	testPublicKey  = "test-public-key"
	testSecret     = "test-secret"
)

type harness struct {
	client *http.Client
	token  string
}

type scenario struct {
	name string
	run  func(h *harness) (string, error)
}

type loginResponse struct {
	TokenValue string `json:"tokenValue"`
}

type createSubAccountRequest struct {
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

type createSubAccountResponse struct {
	AccountID          string                         `json:"accountId"`
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
