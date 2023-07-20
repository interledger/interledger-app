package openpayments

import (
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type PaymentPointer struct {
	ID         string            `db:"id" json:"-"`
	URL        string            `db:"url" json:"id"`
	WalletID   string            `db:"wallet_id" validate:"uuid4" json:"-"`
	Alias      string            `db:"alias" json:"publicName" validate:"required"`
	Asset      currency.Currency `db:"asset" validate:"iso4217"  json:"assetCode"`
	AssetScale int               `db:"scale" validate:"gt=0" json:"assetScale"`
}

type CreateQuoteArgs struct {
	SendPaymentPointer      string `validate:"url"`
	ReceivePaymentPointer   string `json:"receiver" validate:"url"`
	ExpiresAt               time.Time
	SendAmount              currency.Amount `json:"sendAmount"`
	Reference               string
	Description             string
	LinkedAccID             string
	CreatedBy               string // Either the payment pointer from gRPC or the client_id from Openapyments API, which is also a payment pointer
	DestinationIdentity     string
	DestinationIdentityType string `validate:"omitempty,oneof=twitter wallet"`
}

type Quote struct {
	ID                      string          `json:"id"`
	PaymentPointer          string          `json:"paymentPointer"`
	IncomingPayment         string          `json:"receiver"`
	ReceiveAmount           currency.Amount `json:"receiveAmount"`
	SendAmount              currency.Amount `json:"sendAmount"`
	ExpiresAt               time.Time       `json:"expiresAt"`
	CreatedAt               time.Time       `json:"createdAt"`
	FromLinkedAccount       string          `json:"-"`
	CreatedBy               string          `json:"-"`
	DestinationIdentity     string          `json:"-"`
	DestinationIdentityType string          `json:"-"`
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
	CreatedBy          string // Either the payment pointer from gRPC or the client_id from Openapyments API, which is also a payment pointer
}

type IncomingPayment struct {
	ID                 string           `json:"id"`
	PaymentPointer     string           `json:"to"`
	FromPaymentPointer string           `json:"from"`
	IncomingAmount     *currency.Amount `json:"incoming_amount,omitempty"`
	ReceivedAmount     *currency.Amount `json:"outgoing_amount,omitempty"`
	Completed          bool             `json:"completed"`
	ExternalRef        string           `json:"external_ref"`
	Description        string           `json:"description"`
	ExpiresAt          time.Time        `json:"expires_at"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	CreatedBy          string           `json:"-"`
}

type CreateOutgoingPaymentArgs struct {
	QuoteID                 string `json:"quoteId" validate:"required"`
	Description             string `json:"description"`
	IPAddress               string `json:"-" validate:"ip_addr"`
	CreatedBy               string // Either the payment pointer from gRPC or the client_id from Openapyments API, which is also a payment pointer
	GrantID                 string `json:"-"`
	ThreeDSID               string `json:"-"`
	DestinationIdentity     string
	DestinationIdentityType string `validate:"omitempty,oneof=twitter wallet"`
	LinkedAccountTitle      string
}

type OutgoingPayment struct {
	ID                string          `json:"id"`
	PaymentPointer    string          `json:"from"`
	ToPaymentPointer  string          `json:"to"`
	Failed            bool            `json:"failed"`
	Receiver          string          `json:"receiver"`
	SendAmount        currency.Amount `json:"send_amount"`
	ReceiveAmount     currency.Amount `json:"receive_amount"`
	SentAmount        currency.Amount `json:"sent_amount"`
	Description       string          `json:"description"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	FromLinkedAccount string          `json:"-"`
	CreatedBy         string          `json:"-"`
}

type CompleteOutgoingPaymentArgs struct {
	ID              string
	SentAmount      currency.Amount
	IncomingSuccess bool
}

type Jwk struct {
	Kty string `json:"kty,omitempty"`
	E   string `json:"e,omitempty"`
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Use string `json:"use,omitempty"`
}
