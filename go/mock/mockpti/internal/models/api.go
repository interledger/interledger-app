package models

// CreateUserRequest is the request body for POST /users.
type CreateUserRequest struct {
	ID                   string    `json:"id,omitempty"`
	Type                 string    `json:"type,omitempty"`
	DateOfBirth          string    `json:"dateOfBirth,omitempty"`
	Name                 Name      `json:"name,omitempty"`
	Emails               []Email   `json:"emails,omitempty"`
	Addresses            []Address `json:"addresses,omitempty"`
	Phones               []Phone   `json:"phones,omitempty"`
	SourceOfFunds        string    `json:"sourceOfFunds,omitempty"`
	CountryOfCitizenship string    `json:"countryOfCitizenship,omitempty"`
}

// PatchUserRequest is the request body for PATCH /users (merge).
type PatchUserRequest struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	DateOfBirth   string    `json:"dateOfBirth,omitempty"`
	Name          *Name     `json:"name,omitempty"`
	Emails        []Email   `json:"emails,omitempty"`
	Addresses     []Address `json:"addresses,omitempty"`
	Phones        []Phone   `json:"phones,omitempty"`
	SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
}

// StartAssessmentRequest is the request body for POST /users/assessments.
type StartAssessmentRequest struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	DateOfBirth   string    `json:"dateOfBirth,omitempty"`
	Name          Name      `json:"name,omitempty"`
	Emails        []Email   `json:"emails,omitempty"`
	Addresses     []Address `json:"addresses,omitempty"`
	Phones        []Phone   `json:"phones,omitempty"`
	SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
}

// TokenRequest is the request body for POST /auth/jwt.
type TokenRequest struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

// IDResponse is the common response for create operations.
type IDResponse struct {
	ID   string `json:"id"`
	Link string `json:"link,omitempty"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateWalletRequest is the request body for POST /users/{id}/wallets.
type CreateWalletRequest struct {
	ID        string `json:"id,omitempty"`
	Currency  string `json:"currency,omitempty"`
	Type      string `json:"type,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// CreatePaymentInformationRequest is the request body for POST /users/{id}/payment-information.
type CreatePaymentInformationRequest struct {
	Type                  string `json:"type,omitempty"`
	BankAccountNumber     string `json:"bankAccountNumber,omitempty"`
	BankAccountType       string `json:"bankAccountType,omitempty"`
	BankSwiftCode         string `json:"bankSwiftCode,omitempty"`
	BankRoutingNumber     string `json:"bankRoutingNumber,omitempty"`
	BankRoutingCheckDigit string `json:"bankRoutingCheckDigit,omitempty"`
	AccountBankName       string `json:"accountBankName,omitempty"`
}

// TransactionPaymentInformation is a payment method reference inside a transaction request.
type TransactionPaymentInformation struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

// TransactionMethod describes the source or destination of a transaction.
type TransactionMethod struct {
	Currency           string                        `json:"currency,omitempty"`
	PaymentMethodType  string                        `json:"paymentMethodType,omitempty"`
	PaymentInformation TransactionPaymentInformation `json:"paymentInformation,omitempty"`
}

// TransactionInitiator is the user driving the transaction.
type TransactionInitiator struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// CreateDepositRequest is the request body for POST /transactions/deposits.
type CreateDepositRequest struct {
	Initiator         TransactionInitiator `json:"initiator,omitempty"`
	SourceMethod      TransactionMethod    `json:"sourceMethod,omitempty"`
	DestinationMethod TransactionMethod    `json:"destinationMethod,omitempty"`
	Amount            float64              `json:"amount,omitempty"`
	USDAmount         float64              `json:"usdValue,omitempty"`
	Type              string               `json:"type,omitempty"`
	Date              string               `json:"date,omitempty"`
}

// CreateWithdrawalRequest is the request body for POST /transactions/withdrawals.
type CreateWithdrawalRequest struct {
	Initiator         TransactionInitiator `json:"initiator,omitempty"`
	SourceMethod      TransactionMethod    `json:"sourceMethod,omitempty"`
	DestinationMethod TransactionMethod    `json:"destinationMethod,omitempty"`
	Amount            float64              `json:"amount,omitempty"`
	USDAmount         float64              `json:"usdValue,omitempty"`
	Type              string               `json:"type,omitempty"`
	Date              string               `json:"date,omitempty"`
}

// CreateTransferRequest is the request body for POST /transactions/transfers.
type CreateTransferRequest struct {
	Initiator                 TransactionInitiator `json:"initiator,omitempty"`
	SourceTransferMethod      TransactionMethod    `json:"sourceTransferMethod,omitempty"`
	DestinationTransferMethod TransactionMethod    `json:"destinationTransferMethod,omitempty"`
	TransactionGroupID        string               `json:"transactionGroupId,omitempty"`
	Amount                    float64              `json:"amount,omitempty"`
	USDValue                  float64              `json:"usdValue,omitempty"`
	Type                      string               `json:"type,omitempty"`
	Date                      string               `json:"date,omitempty"`
}

// UpdateTransactionRequest is the request body for POST /transactions/{requestId}/updates.
type UpdateTransactionRequest struct {
	TransactionID string `json:"transactionId,omitempty"`
	Feedback      string `json:"feedback,omitempty"`
	ProviderName  string `json:"providerName,omitempty"`
	Payload       string `json:"payload,omitempty"`
}
