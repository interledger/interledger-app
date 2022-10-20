package openpayments

import (
	"context"
	"time"
)

type Client interface {
	CreatePaymentPointer(ctx context.Context, pointer PaymentPointer) (*PaymentPointer, error)
	GetPaymentPointer(ctx context.Context, url string) (*PaymentPointer, error)
	PaymentPointerExists(ctx context.Context, url string) (bool, error)
	ListWalletPaymentPointers(ctx context.Context, walletID string) ([]PaymentPointer, error)
	CreateQuote(ctx context.Context, args CreateQuoteArgs) (Quote, error)
}

type Amount struct {
	Value      uint64 `json:"value,string"`
	AssetCode  string `json:"assetCode"`
	AssetScale int    `json:"assetScale"`
}

type ILPConnection struct {
	ID           string `json:"id"`
	Address      string `json:"ilpAddress"`
	SharedSecret string `json:"sharedSecret"`
	AssetCode    string `json:"assetCode"`
	AssetScale   int    `json:"assetScale"`
}

type IncomingPayment struct {
	ID             string    `json:"id"`
	PaymentPointer string    `json:"paymentPointer"`
	IncomingAmount Amount    `json:"incomingAmount"`
	ReceivedAmount Amount    `json:"receivedAmount"`
	Completed      bool      `json:"completed"`
	ExternalRef    string    `json:"externalRef"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
