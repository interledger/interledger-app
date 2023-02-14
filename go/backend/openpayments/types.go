package openpayments

import (
	"time"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/currency"
)

type PaymentPointer struct {
	ID         string            `db:"id" json:"-"`
	URL        string            `db:"url" json:"id"`
	WalletID   string            `db:"wallet_id" validate:"uuid4" json:"-"`
	Alias      string            `db:"alias" json:"publicName"`
	Asset      currency.Currency `db:"asset" validate:"iso4217"  json:"assetCode"`
	AssetScale int               `db:"scale" validate:"gt=0" json:"assetScale"`
}

type CreateQuoteArgs struct {
	SendPaymentPointer    string `validate:"url"`
	ReceivePaymentPointer string `json:"receiver" validate:"url"`
	ExpiresAt             time.Time
	SendAmount            currency.Amount `json:"sendAmount"`
	Reference             string
	Description           string
	LinkedAccID           string
}

type Quote struct {
	ID                string          `json:"id"`
	PaymentPointer    string          `json:"paymentPointer"`
	IncomingPayment   string          `json:"receiver"`
	ReceiveAmount     currency.Amount `json:"receiveAmount"`
	SendAmount        currency.Amount `json:"sendAmount"`
	ExpiresAt         time.Time       `json:"expiresAt"`
	CreatedAt         time.Time       `json:"createdAt"`
	FromLinkedAccount string          `json:"-"`
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
	IncomingAmount     *currency.Amount
	ExternalRef        string
	ExpiresAt          time.Time
	Description        string
}

type IncomingPayment struct {
	ID                 string           `json:"id"`
	PaymentPointer     string           `json:"paymentPointer"`
	FromPaymentPointer string           `json:"from"`
	IncomingAmount     *currency.Amount `json:"incomingAmount,omitempty"`
	ReceivedAmount     *currency.Amount `json:"receivedAmount,omitempty"`
	Completed          bool             `json:"completed"`
	ExternalRef        string           `json:"externalRef"`
	Description        string           `json:"description"`
	ExpiresAt          time.Time        `json:"expiresAt"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}

type CreateOutgoingPaymentArgs struct {
	IdempotencyKey string `json:"-" validate:"omitempty,uuid"`
	QuoteID        string `json:"quoteId" validate:"required"`
	Description    string `json:"description"`
	ExternalRef    string `json:"externalRef"`
	IPAddress      string `json:"-" validate:"ip_addr"`
}

type OutgoingPayment struct {
	ID                string          `json:"id"`
	PaymentPointer    string          `json:"paymentPointer"`
	ToPaymentPointer  string          `json:"to"`
	Failed            bool            `json:"failed"`
	Receiver          string          `json:"receiver"`
	SendAmount        currency.Amount `json:"sendAmount"`
	ReceiveAmount     currency.Amount `json:"receiveAmount"`
	SentAmount        currency.Amount `json:"sentAmount"`
	Description       string          `json:"description"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	FromLinkedAccount string          `json:"-"`
}

type CompleteOutgoingPaymentArgs struct {
	ID         string
	SentAmount currency.Amount
}

func BaseURL() string {
	if env.IsProd() {
		return "https://open.fynbos.app"
	}
	if env.IsDev() {
		return "https://eu1.open.fynbos.app"
	}
	return "https://eu1.open.fynbos.app"
}
