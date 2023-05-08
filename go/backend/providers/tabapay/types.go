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
}

type PullFromCardArgs struct {
	WalletID    string
	ProviderID  string
	ReferenceID string
	Amount      currency.Amount
	ThreeDSID   string
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
		Amount         currency.Amount
		IdempotencyKey string
		CardID         string
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
		IdempotencyKey          string
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
		IdempotencyKey string
		ThreeDSID      string
		JWT            string
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

func IsSuccessfulTransaction(trx Transaction) bool {
	return (trx.NetworkRC == "00" || trx.NetworkRC == "000") && (trx.Status == string(external.TransactionStatusOk) || trx.Status == string(external.TransactionStatusCompleted))
}

const dictionary = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789@!#$%^&*()<>?{}|;~"

func NewReferenceID() string {
	refBytes := make([]byte, 15)

	for i := range refBytes {
		refBytes[i] = dictionary[rand.Int63()%int64(len(dictionary))]
	}

	return string(refBytes)
}
