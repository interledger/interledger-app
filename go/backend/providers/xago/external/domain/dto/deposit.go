package dto

import "time"

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

type TestDepositRequest struct {
	RunTestDeposit    bool    `json:"runTestDeposit"`
	Amount            float64 `json:"amount"`
	DepositReference  string  `json:"depositReference"`
	BankTransactionID string  `json:"bankTransactionId"`
	CurrencyCode      string  `json:"currencyCode"`
}
