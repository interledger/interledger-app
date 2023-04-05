package tabapay

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

var (
	ProviderName = "tabapay"
	TypeCard     = "card"
)

type CreateCardArgs struct {
	IdempotencyKey string
	WalletID       string
	Name           string
	CardNumber     string
	CVV            string
	ExpirationDate string
}

type PullFromCardArgs struct {
	WalletID    string
	ProviderID  string
	ReferenceID string
	Amount      currency.Amount
}

type PushToCardArgs = PullFromCardArgs

type Transaction struct {
	ID             string
	ReferenceID    string
	Status         string
	OriginalStatus string
	Amount         currency.Amount
	ReversalStatus string
}

type Await func(ctx context.Context, result interface{}) error

type (
	Init3DSArgs struct {
		Amount            currency.Amount
		OutgoingPaymentID string
		CardID            string
	}

	Init3DSResponse struct {
		ID                  string
		JWT                 string
		DeviceCollectionURL string
	}
)

type AuthenticationIndicator string

var (
	AuthenticatorIndicatorPayment AuthenticationIndicator = "01"
	AuthenticatorIndicatorAddCard AuthenticationIndicator = "04"
)

type TransactionMode string

var (
	TransactionModeMobile   TransactionMode = "P"
	TransactionModeComputer TransactionMode = "S"
)

type ProductCode string

var (
	ProductCodeAccountFunding       ProductCode = "ACF"
	ProductCodeQuasiCashTransaction ProductCode = "QCT"
)

type BrowserInfo = external.BrowserInfo
type DeviceChannelType = external.DeviceChannelType

var (
	DeviceChannelBrowser = external.DeviceChannelBrowser
	DeviceChannelSDK     = external.DeviceChannelSDK
)

type (
	Lookup3DSArgs struct {
		OutgoingPaymentID       string
		CardID                  string
		ThreeDSID               string
		AuthenticationIndicator AuthenticationIndicator
		TransactionMode         TransactionMode
		ProductCode             ProductCode
		BrowserInfo             BrowserInfo
		DeviceChannel           DeviceChannelType
		Amount                  currency.Amount
	}

	Lookup3DSResponse struct {
		Version                string
		Enrolled               string
		ProcessorTransactionID string
		DsTransactionID        string
		Status                 string
		ChallengeURL           string
		Payload                string
	}
)
