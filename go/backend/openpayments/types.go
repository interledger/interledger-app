package openpayments

import "time"

type PaymentPointer struct {
	ID         string `db:"id" json:"-"`
	URL        string `db:"url" json:"id"`
	WalletID   string `db:"wallet_id" validate:"uuid4" json:"-"`
	Alias      string `db:"alias" json:"publicName"`
	Asset      string `db:"asset" validate:"iso4217"  json:"assetCode"`
	AssetScale int    `db:"scale" validate:"gt=0" json:"assetScale"`
}

type CreateQuoteArgs struct {
	SendPaymentPointer    string `validate:"url"`
	ReceivePaymentPointer string `json:"receiver" validate:"url"`
	ExpiresAt             time.Time
	SendAmount            Amount `json:"sendAmount"`
	Reference             string
	Description           string
}

type Amount struct {
	Value      uint64 `validate:"gt=0" json:"value,string"`
	Asset      string `validate:"iso4217"  json:"assetCode"`
	AssetScale int    `validate:"gt=0" json:"assetScale"`
}

type Quote struct {
	ID              string    `json:"id"`
	PaymentPointer  string    `json:"paymentPointer"`
	IncomingPayment string    `json:"receiver"`
	ReceiveAmount   Amount    `json:"receiveAmount"`
	SendAmount      Amount    `json:"sendAmount"`
	ExpiresAt       time.Time `json:"expiresAt"`
	CreatedAt       time.Time `json:"createdAt"`
}

type ILPConnection struct {
	ID           string `json:"id"`
	Address      string `json:"ilpAddress"`
	SharedSecret string `json:"sharedSecret"`
	AssetCode    string `json:"assetCode"`
	AssetScale   int    `json:"assetScale"`
}

type CreateIncomingPaymentArgs struct {
	PaymentPointer     string
	FromPaymentPointer string
	IncomingAmount     *Amount
	ExternalRef        string
	ExpiresAt          time.Time
	Description        string
}

type IncomingPayment struct {
	ID                 string    `json:"id"`
	PaymentPointer     string    `json:"paymentPointer"`
	FromPaymentPointer string    `json:"from"`
	IncomingAmount     *Amount   `json:"incomingAmount,omitempty"`
	ReceivedAmount     *Amount   `json:"receivedAmount,omitempty"`
	Completed          bool      `json:"completed"`
	ExternalRef        string    `json:"externalRef"`
	Description        string    `json:"description"`
	ExpiresAt          time.Time `json:"expiresAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type CreateOutgoingPaymentArgs struct {
	QuoteID     string `json:"quoteId" validate:"required"`
	Description string `json:"description"`
	ExternalRef string `json:"externalRef"`
	IPAddress   string `json:"-" validate:"ip_addr"`
}

type OutgoingPayment struct {
	ID               string    `json:"id"`
	PaymentPointer   string    `json:"paymentPointer"`
	ToPaymentPointer string    `json:"to"`
	Failed           bool      `json:"failed"`
	Receiver         string    `json:"receiver"`
	SendAmount       Amount    `json:"sendAmount"`
	ReceiveAmount    Amount    `json:"receiveAmount"`
	SentAmount       Amount    `json:"sentAmount"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type CompleteOutgoingPaymentArgs struct {
	ID         string
	SentAmount Amount
}

type TransactionType string

const (
	TransactionTypeIncomingPayment TransactionType = "incoming"
	TransactionTypeOutgoingPayment TransactionType = "outgoing"
)

// Transaction is abstract information representing either an incoming or outgoing payment.
// This is used for listing transactions for the frontend
type Transaction struct {
	ID          string
	Source      string
	Destination string
	Note        string
	Type        TransactionType
	Timestamp   time.Time
	Amount      Amount
}
