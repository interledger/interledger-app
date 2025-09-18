package external

import (
	"encoding/json"
	"time"
)

const accessTokenID = "ad317668-0e30-4936-8b8d-b2517b2464fd"

var (
	IdentityTypeCorporate  = "corporate"
	IdentityTypeIndividual = "individual"
)

type SubAccount struct {
	AccountID      string                      `json:"accountId,omitempty"`
	DepositAddress string                      `json:"depositAddress,omitempty"`
	DepositTag     int                         `json:"depositTag,omitempty"`
	DepositDetails map[string][]DepositDetails `json:"bankDepositDetails,omitempty"`
	Beneficiaries  []Beneficiaries             `json:"beneficiaries,omitempty"`
}

type DepositDetails struct {
	BankName       string `json:"bankName,omitempty"`
	AccountName    string `json:"accountName,omitempty"`
	AccountNumber  string `json:"accountNumber,omitempty"`
	Iban           string `json:"IBAN,omitempty"`
	BankAddress    string `json:"bankAddress:,omitempty"`
	AccountAddress string `json:"accountAddress,omitempty"`
	SwiftBIC       string `json:"swiftBIC,omitempty"`
	BranchCode     string `json:"branchCode,omitempty"`
}

type Beneficiaries struct {
	BeneficiaryID      string `json:"beneficiaryId,omitempty"`
	BeneficiaryType    string `json:"beneficiaryType,omitempty"`
	CurrencyID         string `json:"currencyId,omitempty"`
	BankName           string `json:"bankName,omitempty"`
	AccountNumber      string `json:"accountNumber,omitempty"`
	AccountName        string `json:"accountName,omitempty"`
	DepositReference   string `json:"depositReference,omitempty"`
	BeneficiaryAction  string `json:"beneficiaryAction,omitempty"`
	DestinationAddress string `json:"destinationAddress,omitempty"`
	DestinationTag     string `json:"destinationTag,omitempty"`
	Reference          string `json:"reference,omitempty"`
}

type AccessToken struct {
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (ac *AccessToken) IsExpired() bool {
	return time.Now().After(ac.ExpiresAt)
}

type SubAccountReq struct {
	FirstName       string `json:"firstName,omitempty"`
	LastName        string `json:"lastName,omitempty"`
	Email           string `json:"email,omitempty"`
	MobileNumber    string `json:"mobileNumber,omitempty"`
	IdentityType    string `json:"identityType,omitempty"`
	IDNumber        string `json:"idNumber,omitempty"`
	PhysicalAddress string `json:"physicalAddress,omitempty"`
	PersonaURL      string `json:"thirdPartyVerificationUrl,omitempty"`
}

type CreateBeneficiaryReq struct {
	Name                       string                      `json:"name,omitempty"`
	Scope                      string                      `json:"scope,omitempty"`
	CurrencyCode               string                      `json:"currencyCode,omitempty"`
	AccountNumber              string                      `json:"accountNumber,omitempty"`
	BranchCode                 string                      `json:"branchCode,omitempty"`
	BankName                   string                      `json:"bankName,omitempty"`
	BankCountry                string                      `json:"bankCountry,omitempty"`
	AccountName                string                      `json:"accountName,omitempty"`
	Reference                  string                      `json:"reference,omitempty"`
	Iban                       string                      `json:"IBAN,omitempty"`
	Bic                        string                      `json:"BIC,omitempty"`
	BeneficiaryPhysicalAddress string                      `json:"beneficiaryPhysicalAddress,omitempty"`
	BeneficiaryDistrict        string                      `json:"beneficiaryDistrict,omitempty"`
	BeneficiaryCity            string                      `json:"beneficiaryCity,omitempty"`
	BeneficiaryCountry         string                      `json:"beneficiaryCountry,omitempty"`
	BeneficiaryPostalCode      string                      `json:"beneficiaryPostalCode,omitempty"`
	BeneficiaryAddress         string                      `json:"beneficiaryAddress,omitempty"`
	AccountType                string                      `json:"accountType,omitempty"`
	MobileNumber               string                      `json:"mobileNumber,omitempty"`
	KYCRequest                 CreateBeneficiaryKYCRequest `json:"kycRequest"`
}

type CreateBeneficiaryKYCRequest struct {
	IsOwn        bool   `json:"isOwn,omitempty"`
	SubAccountID string `json:"existingIdentityId,omitempty"`
}

type CreateBeneficiaryResp struct {
	Status        int                    `json:"status,omitempty"`
	Beneficiaries []AccountBeneficiaries `json:"beneficiaries"`
}

type AccountBeneficiaries struct {
	BranchCode         string          `json:"branchCode"`
	Reference          string          `json:"reference"`
	BeneficiaryAddress string          `json:"beneficiaryAddress"`
	BankName           string          `json:"bankName"`
	AccountNumber      string          `json:"accountNumber"`
	Status             string          `json:"status"`
	CurrencyCode       string          `json:"currencyCode"`
	ID                 string          `json:"uuid"`
	Scope              string          `json:"scope"`
	Name               string          `json:"name"`
	Wallet             json.RawMessage `json:"wallet"`
}

type Pagination struct {
	NumberOfPages int `json:"numberOfPages,omitempty"`
	Limit         int `json:"limit"`
	PageNumber    int `json:"pageNumber"`
}

type ListBeneficiariesResponse struct {
	Pagination    Pagination             `json:"meta,omitempty"`
	Beneficiaries []AccountBeneficiaries `json:"values,omitempty"`
}

type CreateTransferReq struct {
	Amount          float64 `json:"amount,omitempty"`
	CurrencyCode    string  `json:"currencyCode,omitempty"`
	BeneficiaryID   string  `json:"beneficiaryId,omitempty"`
	TransactionType string  `json:"transactionType,omitempty"`
	IdempotencyKey  string  `json:"idempotencyKey,omitempty"`
	Reference       string  `json:"reference"`
}

type Deposit struct {
	IsRequested            bool      `json:"isRequested"`
	IsDuplicate            bool      `json:"isDuplicate"`
	DuplicateTransactionID string    `json:"duplicateTransactionId"`
	OriginAmount           float64   `json:"originAmount"`
	Amount                 float64   `json:"amount"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"createdAt"`
	SettledAt              string    `json:"settledAt"`
	AccountID              string    `json:"accountId"`
	TransactionID          string    `json:"transactionId"`
}

type ListDepositsResponse struct {
	Pagination Pagination `json:"meta,omitempty"`
	Deposits   []Deposit  `json:"data,omitempty"`
}

type Withdrawal struct {
	ID         string  `json:"id"`
	Total      float64 `json:"total"`
	Amount     float64 `json:"amount"`
	Commission float64 `json:"commission"`
	Status     string  `json:"status"`
	Currency   string  `json:"currencyCode"`
}

type TestDepositReq struct {
	RunTestDeposit   bool    `json:"runTestDeposit"`
	Amount           float64 `json:"amount"`
	DepositReference string  `json:"depositReference"`

	// Random UUID for testing
	BankTransactionID string `json:"bankTransactionId"`
	CurrencyCode      string `json:"currencyCode"`
}

type UpdateInquiryLinkRequest struct {
	ThirdPartyVerificationUrl string `json:"thirdPartyVerificationUrl"`
}
