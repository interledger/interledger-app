package external

import (
	"time"

	"github.com/interledger/interledger-app/go/backend/currency"
)

const CardPaymentInformationType = "ENCRYPTED_CREDIT_CARD"

type (
	CreateUserArgs struct {
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
	PutUserArgs = CreateUserArgs

	CreateUserResponse struct {
		ID   string `json:"id,omitempty"`
		Link string `json:"link,omitempty"`
	}

	CreateWalletArgs struct {
		WalletID  string `json:"id,omitempty"` // todo(bradu) renaname walletid -> id
		Currency  string `json:"currency,omitempty"`
		UserID    string `json:"-"`
		Type      string `json:"type,omitempty"`
		Reference string `json:"reference,omitempty"`
	}

	StartUserAssessmentArgs struct {
		ID            string    `json:"id,omitempty"`
		Type          string    `json:"type,omitempty"`
		DateOfBirth   string    `json:"dateOfBirth,omitempty"`
		Name          Name      `json:"name,omitempty"`
		Emails        []Email   `json:"emails,omitempty"`
		Addresses     []Address `json:"addresses,omitempty"`
		Phones        []Phone   `json:"phones,omitempty"`
		SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
		ScenarioID    string    `json:"-"`
	}

	PatchUserArgs struct {
		ID            string    `json:"id,omitempty"`
		Type          string    `json:"type,omitempty"`
		DateOfBirth   string    `json:"dateOfBirth,omitempty"`
		Name          *Name     `json:"name,omitempty"`
		Emails        []Email   `json:"emails,omitempty"`
		Addresses     []Address `json:"addresses,omitempty"`
		Phones        []Phone   `json:"phones,omitempty"`
		SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
	}

	TransferArgs struct {
		RequestID                 string              `json:"-"`
		ScenarioID                string              `json:"-"`
		SessionID                 string              `json:"-"`
		TransactionGroup          string              `json:"transactionGroupId,omitempty"`
		TransactionTotal          Total               `json:"transactionTotal"`
		SubClientID               string              `json:"subClientId,omitempty"`
		USDValue                  float64             `json:"usdValue,omitempty"`
		Amount                    float64             `json:"amount,omitempty"`
		Date                      string              `json:"date,omitempty"`
		Initiator                 User                `json:"initiator"`
		PTIMeta                   map[string]any      `json:"ptiMeta,omitempty"`
		ClientMeta                map[string]any      `json:"clientMeta,omitempty"`
		Type                      string              `json:"type,omitempty"`
		SourceTransferMethod      WalletPaymentMethod `json:"sourceTransferMethod,omitempty"`
		DestinationTransferMethod WalletPaymentMethod `json:"destinationTransferMethod,omitempty"`
		Destination               User                `json:"destination,omitempty"`
		DisableWebhook            bool                `json:"-"`
	}

	TokenArgs struct {
		URL    string `json:"url"`
		Method string `json:"method"`
	}

	TokenResponse struct {
		AccessToken string  `json:"accessToken"`
		ExpiresAt   float64 `json:"expiresAt"`
		TokenType   string  `json:"tokenType"`
	}
)

type (
	Name struct {
		First  string `json:"firstName,omitempty"`
		Last   string `json:"lastName,omitempty"`
		Middle string `json:"middleName,omitempty"`
	}

	Email struct {
		Address string `json:"address,omitempty"`
		Default bool   `json:"default,omitempty"`
	}

	Phone struct {
		Number  string `json:"number,omitempty"`
		Type    string `json:"type,omitempty"`
		Default bool   `json:"default,omitempty"`
	}

	Address struct {
		Street     string `json:"streetAddress,omitempty"`
		City       string `json:"city,omitempty"`
		PostalCode string `json:"postalCode,omitempty"`
		// StateCode  StateCode   `json:"stateCode,omitempty"`
		StateCode string `json:"stateCode,omitempty"`
		// Country   CountryCode `json:"country,omitempty"`
		Country string `json:"country,omitempty"`
		Default bool   `json:"default,omitempty"`
	}

	StateCode struct {
		Code string `json:"code,omitempty"`
	}

	CountryCode struct {
		Code string `json:"code,omitempty"`
	}

	Wallet struct {
		WalletID       string  `json:"walletId,omitempty"`
		Currency       string  `json:"currency,omitempty"`
		Reference      string  `json:"reference,omitempty"`
		CreateDateTime string  `json:"createDateTime,omitempty"`
		Balance        float64 `json:"balance"`
	}

	Assessment struct {
		ResourceType  string `json:"resourceType"`
		ClientID      string `json:"clientId"`
		RequestID     string `json:"requestId"`
		UserId        string `json:"userId"`
		Date          string `json:"date"`
		Assessment    string `json:"assessment"`
		Tier          int    `json:"tier"`
		RefusalReason string `json:"refusalReason"`
	}

	User struct {
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
	}

	PaymentInformation struct {
		ID                string `json:"id,omitempty"`
		Type              string `json:"type"` // BANK_ACCOUNT, ENCRYPTED_CREDIT_CARD, TOKEN, WALLET
		Currency          string `json:"currency,omitempty"`
		BillingEmail      string `json:"billingEmail,omitempty"`
		BankAccountNumber string `json:"bankAccountNumber,omitempty"`
		BankRoutingNumber string `json:"bankRoutingNumber,omitempty"`
		AccountHolderName string `json:"accountHolderName,omitempty"`
		WalletID          string `json:"walletId,omitempty"`
	}

	UpdateTxStatusArgs struct {
		RequestID     string    `json:"-"`
		TransactionID string    `json:"transactionId"`
		Feedback      string    `json:"feedback"`
		Date          time.Time `json:"date"`
		ProviderName  string    `json:"providerName"`
		Payload       string    `json:"payload"`
	}

	Subtotal struct {
		Amount float64 `json:"amount"`
	}

	PaymentTotal struct {
		Subtotal Subtotal `json:"subtotal"`
	}

	StatusPayload struct {
		Status       string       `json:"status"`
		PaymentTotal PaymentTotal `json:"paymentTotal"`
		ProviderName string       `json:"providerName"`
	}

	DepositArgs struct {
		RequestID                 string `json:"-"`
		ScenarioID                string `json:"-"`
		UserID                    string `json:"-"`
		SessionID                 string `json:"-"`
		ExternalWalletID          string
		ExternalPaymentMethodID   string
		ExternalPaymentMethodType string // card or bank
		Amount                    currency.Amount
		AccountHolderName         string
	}
	WithdrawalArgs struct {
		RequestID             string `json:"-"`
		ScenarioID            string `json:"-"`
		UserID                string `json:"-"`
		SessionID             string `json:"-"`
		ExternalWalletID      string
		ExternalBankAccountID string
		Amount                currency.Amount
		AccountHolderName     string
	}

	CreateTxResponse struct {
		ID   string `json:"id,omitempty"`
		Link string `json:"link,omitempty"`
	}

	WithdrawDetails struct {
		ID               string
		UserID           string
		Amount           currency.Amount
		ExternalWalletID string
	}

	InternalCreateWithdrawalArgs struct {
		Initiator         Initiator                   `json:"initiator,omitempty"`
		SourceMethod      SourceMethod                `json:"sourceMethod,omitempty"`
		DestinationMethod WithdrawalDestinationMethod `json:"destinationMethod,omitempty"`
		Amount            float64                     `json:"amount,omitempty"`
		USDAmount         float64                     `json:"usdValue,omitempty"`
		Type              string                      `json:"type,omitempty"`
		Date              string                      `json:"date,omitempty"`
	}

	WithdrawalDestinationMethod struct {
		Currency           string             `json:"currency,omitempty"`
		PaymentMethodType  string             `json:"paymentMethodType,omitempty"`
		PaymentInformation PaymentInformation `json:"paymentInformation,omitempty"`
	}

	internalCreateDepositArgs struct {
		Initiator         User              `json:"initiator,omitempty"`
		SourceMethod      SourceMethod      `json:"sourceMethod,omitempty"`
		DestinationMethod DestinationMethod `json:"destinationMethod,omitempty"`
		Amount            float64           `json:"amount,omitempty"`
		USDAmount         float64           `json:"usdValue,omitempty"`
		Type              string            `json:"type,omitempty"`
		Date              string            `json:"date,omitempty"`
	}
	Initiator struct {
		UserID string `json:"id,omitempty"`
		Type   string `json:"type,omitempty"`
	}
	SourceMethod struct {
		Currency           string             `json:"currency,omitempty"`
		PaymentInformation PaymentInformation `json:"paymentInformation,omitempty"`
		PaymentMethodType  string             `json:"paymentMethodType,omitempty"`
	}
	DestinationInformation struct {
		Type string `json:"type,omitempty"`
		ID   string `json:"id,omitempty"`
	}
	DestinationMethod struct {
		PaymentMethodType  string                 `json:"paymentMethodType,omitempty"`
		PaymentInformation DestinationInformation `json:"paymentInformation,omitempty"`
	}

	WalletPaymentMethod struct {
		PaymentMethodType  string     `json:"paymentMethodType,omitempty"`
		PaymentInformation WalletType `json:"paymentInformation"`
	}

	WalletType struct {
		ID   string `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	}

	TransactionAssessment struct {
		ResourceType                      string                            `json:"resourceType"`
		RequestID                         string                            `json:"requestId"`
		ClientID                          string                            `json:"clientId"`
		UserID                            string                            `json:"userId"`
		Date                              string                            `json:"date"`
		Assessment                        string                            `json:"assessment"`
		Risk                              string                            `json:"risk"`
		Amount                            float64                           `json:"amount"`
		TransactionType                   string                            `json:"transactionType"`
		Meta                              map[string]any                    `json:"meta"`
		TransactionMonitoringResultDetail TransactionMonitoringResultDetail `json:"transactionMonitoringResultDetail"`
	}

	TransactionMonitoringResultDetail struct {
		ComplianceProviderResponseCode string `json:"complianceProviderResponseCode"`
	}

	IDResponse struct {
		ID   string `json:"id"`
		Link string `json:"link"`
	}

	Cost struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}

	PaymentStatusDetail struct {
		ProviderResponseCode     string `json:"providerResponseCode"`
		ProviderResponseCategory string `json:"providerResponseCategory"`
	}

	Total struct {
		Fee      Cost `json:"fee"`
		Total    Cost `json:"total"`
		Subtotal Cost `json:"subtotal"`
	}

	TransactionStatus struct {
		ResourceType        string              `json:"resourceType"`
		RequestID           string              `json:"requestId"`
		ClientID            string              `json:"clientId"`
		UserID              string              `json:"userId"`
		Date                string              `json:"date"`
		Status              string              `json:"status"`
		TransactionType     string              `json:"transactionType"`
		PaymentMethod       string              `json:"paymentMethod"`
		PaymentStatusDetail PaymentStatusDetail `json:"paymentStatusDetail"`
		Amount              float64             `json:"amount"`
		BillingEmail        string              `json:"billingEmail"`
		Total               Total               `json:"total"`
		Currency            string              `json:"currency"`
		AdditionalInfos     map[string]any      `json:"additionalInfos"`
	}

	ExternalPaymentInformation struct {
		ID   string `json:"id,omitempty"`
		Type string `json:"type,omitempty"`
	}

	EncryptedCreditCardPaymentInformation struct {
		ID                  string   `json:"id,omitempty"`
		Type                string   `json:"type,omitempty"`
		CreditCardLast4     string   `json:"creditCardLast4,omitempty"`
		CreditCardType      string   `json:"creditCardType,omitempty"`
		CreditCardBin       string   `json:"creditCardBin,omitempty"`
		CreditCardReference string   `json:"creditCardReference,omitempty"`
		CreditCardAddress   *Address `json:"creditCardAddress,omitempty"`
		ExpirationYear      string   `json:"expirationYear,omitempty"`
		ExpirationMonth     string   `json:"expirationMonth,omitempty"`
		CardHolderFirstName string   `json:"cardHolderFirstName,omitempty"`
		CardHolderLastName  string   `json:"cardHolderLastName,omitempty"`
	}

	BankAccountPaymentInformation struct {
		ID                    string `json:"id,omitempty"`
		Type                  string `json:"type,omitempty"`
		BankAccountNumner     string `json:"bankAccountNumber,omitempty"`
		BankAccountType       string `json:"bankAccountType,omitempty"`
		BankSwiftCode         string `json:"bankSwiftCode,omitempty"`
		BankRoutingNumber     string `json:"bankRoutingNumber,omitempty"`
		BankRoutingCheckDigit string `json:"bankRoutingCheckDigit,omitempty"`
		AccountBankName       string `json:"accountBankName,omitempty"`
	}

	TokenPaymentInformation struct {
		ID                string `json:"id,omitempty"`
		Type              string `json:"type,omitempty"`
		TokenAddress      string `json:"tokenAddress,omitempty"`
		TokenType         string `json:"tokenType,omitempty"`
		Blockchain        string `json:"blockchain,omitempty"`
		PrivateBlockchain bool   `json:"privateBlockchain,omitempty"`
	}
)

type SignatureBase struct {
	Method      string
	Payload     []byte
	ContentType string
	Date        time.Time
	ClientID    string
	Path        string
}
