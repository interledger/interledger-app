package dto

type CreateTransferRequest struct {
	Amount          float64 `json:"amount,omitempty"`
	CurrencyCode    string  `json:"currencyCode,omitempty"`
	BeneficiaryID   string  `json:"beneficiaryId,omitempty"`
	TransactionType string  `json:"transactionType,omitempty"`
	IdempotencyKey  string  `json:"idempotencyKey,omitempty"`
	Reference       string  `json:"reference"`
}
