package models

import "time"

// User represents a PTI user.
type User struct {
	ID                 string         `json:"id,omitempty"`
	Type               string         `json:"type,omitempty"`
	Status             string         `json:"status,omitempty"`
	StatusReason       string         `json:"statusReason,omitempty"`
	Tags               []string       `json:"tags,omitempty"`
	SourceOfFunds      string         `json:"sourceOfFunds,omitempty"`
	UserCreationDate   string         `json:"userCreateionDate,omitempty"`
	Addresses          []Address      `json:"addresses,omitempty"`
	UserPTIMetaData    map[string]any `json:"userPtiMeta,omitempty"`
	UserClientMetaData map[string]any `json:"userClientMeta,omitempty"`
	Emails             []Email        `json:"emails,omitempty"`
	Phones             []Phone        `json:"phones,omitempty"`
	Name               *Name          `json:"name,omitempty"`
	DateOfBirth        string         `json:"dateOfBirth,omitempty"`
	CreatedAt          time.Time      `json:"-"`
}

// Name represents a person's name.
type Name struct {
	First  string `json:"firstName,omitempty"`
	Last   string `json:"lastName,omitempty"`
	Middle string `json:"middleName,omitempty"`
}

// Email represents an email address.
type Email struct {
	Address string `json:"address,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Phone represents a phone number.
type Phone struct {
	Number  string `json:"number,omitempty"`
	Type    string `json:"type,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Address represents a physical address.
type Address struct {
	Street     string `json:"streetAddress,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	StateCode  string `json:"stateCode,omitempty"`
	Country    string `json:"country,omitempty"`
	Default    bool   `json:"default,omitempty"`
}

// Assessment represents a PTI user KYC assessment.
type Assessment struct {
	ResourceType  string `json:"resourceType"`
	ClientID      string `json:"clientId"`
	RequestID     string `json:"requestId"`
	UserID        string `json:"userId"`
	Date          string `json:"date"`
	Assessment    string `json:"assessment"`
	Tier          int    `json:"tier"`
	RefusalReason string `json:"refusalReason,omitempty"`
}

// Wallet represents a PTI user wallet.
type Wallet struct {
	WalletID       string  `json:"walletId,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	Reference      string  `json:"reference,omitempty"`
	CreateDateTime string  `json:"createDateTime,omitempty"`
	Balance        float64 `json:"balance"`
	UserID         string  `json:"-"`
}

// PaymentInformation represents a stored payment method for a PTI user.
type PaymentInformation struct {
	ID                    string `json:"id,omitempty"`
	Type                  string `json:"type"`
	BankAccountNumber     string `json:"bankAccountNumber,omitempty"`
	BankAccountType       string `json:"bankAccountType,omitempty"`
	BankSwiftCode         string `json:"bankSwiftCode,omitempty"`
	BankRoutingNumber     string `json:"bankRoutingNumber,omitempty"`
	BankRoutingCheckDigit string `json:"bankRoutingCheckDigit,omitempty"`
	AccountBankName       string `json:"accountBankName,omitempty"`
	UserID                string `json:"-"`
}

// TokenResponse represents a PTI JWT token response.
type TokenResponse struct {
	AccessToken string  `json:"accessToken"`
	ExpiresAt   float64 `json:"expiresAt"`
	TokenType   string  `json:"tokenType"`
}

// Transaction represents a PTI transaction (deposit, withdrawal, or transfer).
type Transaction struct {
	RequestID       string    `json:"requestId"`
	Status          string    `json:"status"`
	TransactionType string    `json:"transactionType"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Date            string    `json:"date"`
	UserID          string    `json:"userId,omitempty"`
	ResourceType    string    `json:"resourceType"`
	ClientID        string    `json:"clientId,omitempty"`
	CreatedAt       time.Time `json:"-"`
}

// TransactionUpdate represents feedback sent back to PTI for a transaction.
type TransactionUpdate struct {
	ID            string    `json:"id"`
	RequestID     string    `json:"-"`
	TransactionID string    `json:"transactionId"`
	Feedback      string    `json:"feedback"`
	Date          time.Time `json:"date"`
	ProviderName  string    `json:"providerName"`
	Payload       string    `json:"payload"`
}

// Job represents a unit of async work (webhook delivery, etc.)
type Job struct {
	ID          string                 `json:"id"`
	JobType     string                 `json:"job_type"`
	Data        map[string]interface{} `json:"data"`
	Attempts    int                    `json:"attempts"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	NotBefore   time.Time              `json:"not_before"`
	LastError   string                 `json:"last_error"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}
