package models

import "time"

// Card-related API and storage models

type Pagination struct {
	PageNumber uint `json:"pageNumber"`
	PageSize   uint `json:"pageSize"`
	TotalPages uint `json:"totalPages"`
}

type ListCardsResponse struct {
	Data       []Card     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Customer struct {
	ID        *string                   `json:"id"`
	SourceID  string                    `json:"sourceId"`
	Type      string                    `json:"type"`
	Code      string                    `json:"code"`
	TaxNumber string                    `json:"taxNumber"`
	KYCStatus string                    `json:"kycStatus"`
	Addresses []CustomerDeliveryAddress `json:"addresses"`
	Accounts  []Account                 `json:"accounts"`
	CreatedAt time.Time                 `json:"createdAt"`
}

type CustomerResponse struct {
	WalletAddress string   `json:"walletAddress"`
	Customer      Customer `json:"customers"`
}

type Account struct {
	ID               *string   `json:"id"`
	SourceID         string    `json:"sourceId"`
	CustomerID       *string   `json:"customerId"`
	CustomerSourceID string    `json:"customerSourceId"`
	ProductCode      string    `json:"productCode"`
	Currency         string    `json:"currency"`
	AccountNumber    string    `json:"accountNumber"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	StatusReasonCode string    `json:"statusReasonCode"`
	Cards            []Card    `json:"cards"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Card struct {
	ID                         string    `json:"id"`
	SourceID                   string    `json:"sourceId"`
	AccountID                  string    `json:"accountId"`
	AccountSourceID            string    `json:"accountSourceId"`
	CustomerID                 string    `json:"customerId"`
	CustomerSourceID           string    `json:"customerSourceId"`
	NameOnCard                 string    `json:"nameOnCard"`
	ProductCode                string    `json:"productCode"`
	PanToken                   string    `json:"panToken"`
	MaskedPan                  string    `json:"maskedPan"`
	Status                     string    `json:"status"`
	StatusReasonCode           *string   `json:"statusReasonCode"`
	LockLevel                  *string   `json:"lockLevel"`
	RelationType               string    `json:"relationType"`
	MembershipFeeEffectiveDate *string   `json:"membershipFeeEffectiveDate"`
	IsFirstTimeLock            bool      `json:"isFirstTimeLock"`
	PlasticCreated             bool      `json:"plasticCreated"`
	OrderPlasticOnSync         bool      `json:"OrderPlasticOnSync"`
	CreatedAt                  time.Time `json:"createdAt"`
}

type CustomerDeliveryAddress struct {
	ID               string  `json:"id"`
	SourceID         string  `json:"sourceId"`
	CustomerID       string  `json:"customerId"`
	CustomerSourceID string  `json:"customerSourceId"`
	Type             string  `json:"type"`
	Line1            string  `json:"line1"`
	Line2            *string `json:"line2"`
	Line3            *string `json:"line3"`
	City             string  `json:"city"`
	PostOffice       *string `json:"postOffice"`
	ZipCode          string  `json:"zipCode"`
	CountryCode      string  `json:"countryCode"`
	Status           string  `json:"status"`
}

type NewCardArgs struct {
	ProductCode string `json:"productCode"`
}

type CardAccount struct {
	ProductCode string      `json:"productCode"`
	Currency    string      `json:"currency"`
	Card        NewCardArgs `json:"card"`
}

type CreateCustomerAndCardArgs struct {
	WalletAddress string                             `json:"walletAddress"`
	Account       CardAccount                        `json:"account"`
	Delivery      *CreateCustomerDeliveryAddressArgs `json:"delivery,omitempty"`
	NameOnCard    string                             `json:"nameOnCard"`
}

type CreateCustomerDeliveryAddressArgs struct {
	Type        string  `json:"type"`
	CountryCode string  `json:"countryCode"`
	Line1       string  `json:"line1"`
	Line2       *string `json:"line2,omitempty"`
	Line3       *string `json:"line3,omitempty"`
	City        string  `json:"city"`
	PostOffice  *string `json:"postOffice,omitempty"`
	ZipCode     string  `json:"zipCode"`
	Reason      string  `json:"reason"`
}

type OrderCardArgs struct {
	Currency          string      `json:"currency"`
	ProductCode       string      `json:"productCode"`
	NameOnCard        string      `json:"nameOnCard"`
	DeliveryAddressID *string     `json:"deliveryAddressId,omitempty"`
	WalletAddress     string      `json:"walletAddress"`
	Card              NewCardArgs `json:"card"`
}

type FreezeCardArgs struct {
	ReasonCode string  `json:"-"`
	Note       *string `json:"note,omitempty"`
}

type UnfreezeCardArgs struct {
	Note *string `json:"note,omitempty"`
}

type CloseCardArgs struct {
	ReasonCode string `json:"-"`
}

type GetCardTokenArgs struct {
	CardID    string  `json:"cardId"`
	PublicKey *string `json:"publicKey,omitempty"`
}

type ChangePinArgs struct {
	Cypher string `json:"cypher"`
}

type CardLimit struct {
	Type       string  `json:"type"`
	Limit      float64 `json:"limit"`
	Currency   string  `json:"currency"`
	IsDisabled bool    `json:"isDisabled"`
}

type CardTokenLink struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

type CardTokenResponse struct {
	Token string          `json:"token"`
	Links []CardTokenLink `json:"links"`
}

type CardTransactionsPagination struct {
	PageNumber   uint `json:"pageNumber"`
	PageSize     uint `json:"pageSize"`
	TotalPages   uint `json:"totalPages"`
	TotalRecords uint `json:"totalRecords"`
}

type CardTransactionsResponse struct {
	Data       []CardTransaction          `json:"data"`
	Pagination CardTransactionsPagination `json:"pagination"`
}

type CreateCardTransactionArgs struct {
	CardID          string  `json:"cardId"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	Type            int     `json:"type"`
	MerchantName    *string `json:"merchantName,omitempty"`
	MerchantCity    *string `json:"merchantCity,omitempty"`
	MerchantCountry *string `json:"merchantCountry,omitempty"`
}

type CardTransaction struct {
	VaultID                   *int    `json:"vaultId"`
	CardID                    *int    `json:"cardId"`
	TransactionID             string  `json:"transactionId"`
	RefTransactionID          *string `json:"refTransactionId"`
	ResponseCode              *string `json:"responseCode"`
	GHResponseCode            string  `json:"ghResponseCode"`
	GHResponseDescription     string  `json:"ghResponseDescription"`
	TransactionAmount         *string `json:"transactionAmount"`
	TransactionCurrency       *string `json:"transactionCurrency"`
	BillingAmount             *string `json:"billingAmount"`
	BillingCurrency           *string `json:"billingCurrency"`
	IsTrxAmountConverted      bool    `json:"isTrxAmountConverted"`
	TerminalID                string  `json:"terminalId"`
	CardScheme                int     `json:"cardScheme"`
	Type                      int     `json:"type"`
	Source                    *string `json:"source"`
	EntryMode                 *int    `json:"entryMode"`
	AuthorizationCode         *string `json:"authorizationCode"`
	SplitIntoMultiple         *bool   `json:"splitIntoMultiple"`
	Authentication            *int    `json:"authentication"`
	AuthenticationProtocol    *int    `json:"authenticationProtocol"`
	Wallet                    *int    `json:"wallet"`
	TransactionDateTime       *string `json:"transactionDateTime"`
	ProcessDateTime           *string `json:"processDateTime"`
	CreatedAt                 string  `json:"createdAt"`
	TxStatus                  *string `json:"txStatus"`
	Operation                 int     `json:"operation"`
	Mcc                       *string `json:"mcc"`
	MerchantStreet            *string `json:"merchantStreet"`
	MerchantCity              *string `json:"merchantCity"`
	MerchantZip               *string `json:"merchantZip"`
	MerchantCountry           *string `json:"merchantCountry"`
	MerchantName              *string `json:"merchantName"`
	MerchantID                *string `json:"merchantId"`
	TransactionClassification *string `json:"transactionClassification"`
	SpendExchangeRate         *string `json:"spendExchangeRate"`
	SpendCurrency             *string `json:"spendCurrency"`
}

// 3DS-related models

type PendingThreeDSConfirmation struct {
	TransactionID    string `json:"transactionId"`
	MerchantName     string `json:"merchantName"`
	PurchaseAmount   string `json:"purchaseAmount"`
	PurchaseCurrency string `json:"purchaseCurrency"`
	PurchaseDate     string `json:"purchaseDate"`
	Timeout          string `json:"timeout"` // ISO 8601 timestamp
}

type ThreeDSPaymentConfirmationArgs struct {
	TransactionID string `json:"-"`          // URL param
	Confirmed     bool   `json:"confirmed"`  // true = approve, false = decline
	AuthMethod    string `json:"authMethod"` // "biometric" | "pin" | "password"
}

type ThreeDSChallenge struct {
	TransactionID    string    `json:"transactionId"`
	CardID           string    `json:"cardId"`
	UserID           string    `json:"userId"`
	MerchantName     string    `json:"merchantName"`
	PurchaseAmount   string    `json:"purchaseAmount"`
	PurchaseCurrency string    `json:"purchaseCurrency"`
	PurchaseDate     string    `json:"purchaseDate"`
	Timeout          time.Time `json:"timeout"`
	Status           string    `json:"status"` // "pending" | "approved" | "declined" | "expired"
	CreatedAt        time.Time `json:"createdAt"`
}
