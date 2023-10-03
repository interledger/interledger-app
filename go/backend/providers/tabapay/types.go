package tabapay

import (
	"context"
	"math/rand"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/env"
)

var (
	ProviderName = "tabapay"
	TypeCard     = "card"
)

const (
	ThreeDSFullyAuthenticated      = "05|02"
	ThreeDSAttemptedAuthentication = "01|06"
	ThreeDSNotSecure               = "00|07"
)

type CreateCardArgs struct {
	WalletID           string
	BasisTheoryTokenID string
	TabapayReferenceID string
}

type PullFromCardArgs struct {
	WalletID       string
	ProviderID     string
	ReferenceID    string
	Amount         currency.Amount
	ThreeDSID      string
	SoftDescriptor string
}

type PushToCardArgs = PullFromCardArgs

type Fees struct {
	Tabapay     string
	Interchange string
	Network     string
}

type Transaction struct {
	ID                 string
	ReferenceID        string
	Network            string
	NetworkRC          string
	Status             string
	OriginalStatus     string
	Amount             currency.Amount
	Fees               Fees
	ReversalStatus     string
	ReversalNetworkRC  string
	ReversalNetworkRC2 string
	ReversalError      string
}

type Await func(ctx context.Context, result interface{}) error

type (
	Init3DSArgs struct {
		Amount  currency.Amount
		OrderID string
		CardID  string
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
type BrowserInfoFields = external.BrowserInfoFields
type DeviceChannelType = external.DeviceChannelType

var NewBrowserInfo = external.NewBrowserInfo

type ColorDepth = external.ColorDepth

func GetColorDepth(depth string) ColorDepth {
	colorDepth := ColorDepth(depth)
	switch colorDepth {
	case external.ColorDepth1:
		return external.ColorDepth1
	case external.ColorDepth4:
		return external.ColorDepth4
	case external.ColorDepth8:
		return external.ColorDepth8
	case external.ColorDepth15:
		return external.ColorDepth15
	case external.ColorDepth16:
		return external.ColorDepth16
	case external.ColorDepth24:
		return external.ColorDepth24
	case external.ColorDepth32:
		return external.ColorDepth32
	case external.ColorDepth48:
		return external.ColorDepth48
	default:
		return external.ColorDepth32
	}
}

var (
	DeviceChannelBrowser = external.DeviceChannelBrowser
	DeviceChannelSDK     = external.DeviceChannelSDK
)

type (
	Lookup3DSArgs struct {
		OrderID                 string
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
		ECI                    string
		UCAF                   string
		XID                    string
		ChallengeURL           string
		Payload                string
	}
)

type (
	Authenticate3DSArgs struct {
		ThreeDSID string
		JWT       string
	}

	Authenticate3DSResponse struct {
		Version3DS             string
		Enrolled               string
		ProcessorTransactionID string
		DsTransactionID        string
		Status                 string
		ECI                    string
		UCAF                   string
		XID                    string
	}
)

type ThreeDSSession struct {
	ID                     string
	CardID                 string
	OrderID                string
	Revision               int
	Amount                 uint64
	Currency               string
	Version                string
	Enrolled               string
	ProcessorTransactionID string
	DsTransactionID        string
	Status                 string
	ECI                    string
	UCAF                   string
	XID                    string
	ChallengeURL           string
	Payload                string
	InitAt                 time.Time
	LookupAt               time.Time
	AuthenticatedAt        time.Time
}

func GetSongbirdURL() string {
	if env.IsProd() {
		return "https://songbird.cardinalcommerce.com/edge/v1/songbird.js"
	}

	return "https://songbirdstag.cardinalcommerce.com/edge/v1/songbird.js"
}

func IsFrictionlessAuthentication(lookup Lookup3DSResponse) bool {
	return lookup.ChallengeURL == ""
}

func IsTransactionStatusUnknown(trx Transaction) bool {
	return trx.Status == string(external.TransactionStatusUnknown)
}

func IsSuccessfulTransaction(trx Transaction) bool {
	return (trx.NetworkRC == "00" || trx.NetworkRC == "000") && (trx.Status == string(external.TransactionStatusOk) || trx.Status == string(external.TransactionStatusCompleted))
}

const dictionary = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

func NewReferenceID() string {
	refBytes := make([]byte, 15)

	for i := range refBytes {
		refBytes[i] = dictionary[rand.Int63()%int64(len(dictionary))]
	}

	return string(refBytes)
}

type FXRate struct {
	Currency      currency.Currency `db:"currency_code"`
	VisaRate      NetworkRate
	MatercardRate NetworkRate
}

type NetworkRate struct {
	BuyRate     float64 `db:"buy_rate"`
	BuyRateInv  float64 `db:"buy_rate_inverted"`
	SellRate    float64 `db:"sell_rate"`
	SellRateInv float64 `db:"sell_rate_inverted"`
}

// FromUSD takes the USD amount we want to convert and returns the amount of the currency we would get.
func (f *NetworkRate) FromUSD(amt float64) float64 {
	return amt * f.BuyRateInv
}

// ToUSD takes the amt of the target currency and converts it to it's USD equivalent.
func (f *NetworkRate) ToUSD(amt float64) float64 {
	return amt * f.SellRate
}
