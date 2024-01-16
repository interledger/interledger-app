package external

import (
	"time"
)

type (
	CreateUserArgs struct {
		ID            string    `json:"id,omitempty"`
		Type          string    `json:"type,omitempty"`
		DateOfBirth   string    `json:"dateOfBirth,omitempty"`
		Name          Name      `json:"name,omitempty"`
		Emails        []Email   `json:"emails,omitempty"`
		Addresses     []Address `json:"addresses,omitempty"`
		Phones        []Phone   `json:"phones,omitempty"`
		SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
	}
	PutUserArgs = CreateUserArgs

	CreateUserResponse struct {
		ID   string `json:"id,omitempty"`
		Link string `json:"link,omitempty"`
	}

	CreateWalletArgs struct {
		UserID    string `json:"-"`
		WalletID  string `json:"walletId,omitempty"`
		Currency  string `json:"currency,omitempty"`
		Reference string `json:"reference,omitempty"`
		Type      string `json:"type,omitempty"`
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
		RequestID                 string                 `json:"-"`
		ScenarioID                string                 `json:"-"`
		TransactionGroup          string                 `json:"transactionGroupId"`
		SubClientID               string                 `json:"subClientId"`
		USDValue                  float64                `json:"usdValue"`
		Amount                    float64                `json:"amount"`
		Date                      string                 `json:"date"`
		Initiator                 User                   `json:"initiator"`
		PTIMeta                   map[string]interface{} `json:"ptiMeta"`
		ClientMeta                map[string]interface{} `json:"clientMeta"`
		Type                      string                 `json:"type"`
		SourceTransferMethod      WalletPaymentMethod    `json:"sourceTransferMethod"`
		DestinationTransferMethod WalletPaymentMethod    `json:"destinationTransferMethod"`
		Destination               *User                  `json:"destination"`
		DisableWebhook            bool                   `json:"-"`
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
		Street     string      `json:"streetAddress,omitempty"`
		City       string      `json:"city,omitempty"`
		PostalCode string      `json:"postalCode,omitempty"`
		StateCode  StateCode   `json:"stateCode,omitempty"`
		Country    CountryCode `json:"country,omitempty"`
		Default    bool        `json:"default,omitempty"`
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
		ID                 string                 `json:"id,omitempty"`
		Type               string                 `json:"type,omitempty"`
		Status             string                 `json:"status,omitempty"`
		StatusReason       string                 `json:"statusReason,omitempty"`
		Tags               []string               `json:"tags,omitempty"`
		PaymentInformation PaymentInformation     `json:"paymentInformation,omitempty"`
		SourceOfFunds      string                 `json:"sourceOfFunds,omitempty"`
		UserCreationDate   string                 `json:"userCreateionDate,omitempty"`
		Addresses          []Address              `json:"addresses,omitempty"`
		UserPTIMetaData    map[string]interface{} `json:"userPtiMeta,omitempty"`
		UserClientMetaData map[string]interface{} `json:"userClientMeta,omitempty"`
		Emails             []Email                `json:"emails,omitempty"`
		Phones             []Phone                `json:"phones,omitempty"`
		Name               Name                   `json:"name,omitempty"`
		DateOfBirth        string                 `json:"dateOfBirth,omitempty"`
	}

	PaymentInformation struct {
		Type                            string                          `json:"paymentInformationType"`
		BankAccountPaymentInformation   BankAccountPaymentInformation   `json:"bankAccountPaymentInformation"`
		EncryptedCardPaymentInformation EncryptedCardPaymentInformation `json:"encryptedCardPaymentInformation"`
	}

	EncryptedCardPaymentInformation struct {
		CreditCardNumberHash string `json:"creditCardNumberHash,omitempty"`
	}

	BankAccountPaymentInformation struct {
		BankAccountNumber string `json:"bankAccountNumber,omitempty"`
	}

	WalletPaymentMethod struct {
		ID   string
		Type string
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
		Meta                              map[string]interface{}            `json:"meta"`
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
		ResourceType        string                 `json:"resourceType"`
		RequestID           string                 `json:"requestId"`
		ClientID            string                 `json:"clientId"`
		UserID              string                 `json:"userId"`
		Date                string                 `json:"date"`
		Status              string                 `json:"status"`
		TransactionType     string                 `json:"transactionType"`
		PaymentMethod       string                 `json:"paymentMethod"`
		PaymentStatusDetail PaymentStatusDetail    `json:"paymentStatusDetail"`
		Amount              float64                `json:"amount"`
		BillingEmail        string                 `json:"billingEmail"`
		Total               Total                  `json:"total"`
		Currency            string                 `json:"currency"`
		AdditionalInfos     map[string]interface{} `json:"additionalInfos"`
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
