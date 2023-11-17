package external

import (
	"encoding/json"
	"time"
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
	Token     string
	ExpiresAt time.Time
}

func (ac *AccessToken) IsExpired() bool {
	return time.Now().After(ac.ExpiresAt)
}

type SubAccountReq struct {
	FirstName                  string `json:"firstName,omitempty"`
	LastName                   string `json:"lastName,omitempty"`
	Email                      string `json:"email,omitempty"`
	MobileNumber               string `json:"mobileNumber,omitempty"`
	IdentificationDocumentType string `json:"identificationDocumentType,omitempty"`
	IdentificationNumber       string `json:"identificationNumber,omitempty"`
	Address                    string `json:"address,omitempty"`
	City                       string `json:"city,omitempty"`
	District                   string `json:"district,omitempty"`
	PostalCode                 string `json:"postalCode,omitempty"`
	AddressDocumentType        string `json:"addressDocumentType,omitempty"`
	Country                    string `json:"country,omitempty"`
	Nationality                string `json:"nationality,omitempty"`
	DateOfBirth                int    `json:"dateOfBirth,omitempty"`
}

type CreateBeneficiaryReq struct {
	Name                       string `json:"name,omitempty"`
	Scope                      string `json:"scope,omitempty"`
	CurrencyCode               string `json:"currencyCode,omitempty"`
	AccountNumber              string `json:"accountNumber,omitempty"`
	BranchCode                 string `json:"branchCode,omitempty"`
	BankName                   string `json:"bankName,omitempty"`
	BankCountry                string `json:"bankCountry,omitempty"`
	AccountName                string `json:"accountName,omitempty"`
	BankBeneficiaryType        string `json:"bankBeneficiaryType,omitempty"`
	Reference                  string `json:"reference,omitempty"`
	Iban                       string `json:"IBAN,omitempty"`
	Bic                        string `json:"BIC,omitempty"`
	BeneficiaryPhysicalAddress string `json:"beneficiaryPhysicalAddress,omitempty"`
	BeneficiaryDistrict        string `json:"beneficiaryDistrict,omitempty"`
	BeneficiaryCity            string `json:"beneficiaryCity,omitempty"`
	BeneficiaryCountry         string `json:"beneficiaryCountry,omitempty"`
	BeneficiaryPostalCode      string `json:"beneficiaryPostalCode,omitempty"`
	BeneficiaryAddress         string `json:"beneficiaryAddress,omitempty"`
	AccountType                string `json:"accountType,omitempty"`
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

type CreateTransactionReq struct {
	Values []TransactionValues `json:"values,omitempty"`
}

type TransactionValues struct {
	Amount          float64 `json:"amount,omitempty"`
	CurrencyCode    string  `json:"currencyCode,omitempty"`
	BeneficiaryID   string  `json:"beneficiaryId,omitempty"`
	TransactionType string  `json:"transactionType,omitempty"`
	IdempotencyKey  string  `json:"idempotencyKey,omitempty"`
}
