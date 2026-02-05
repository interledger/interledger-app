package gatehub

import (
	"database/sql"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/wallets"
)

// Config holds the configuration for the Gatehub provider
type Config struct {
	AppID                  string
	Secret                 string
	CardAppID              string
	GatewayID              string
	CardAccountProductCode string
	PaywiserEuroVaultID    string
	SendingUserID          string
	SendingUserAddress     string
	WebhookSecret          string
	FallbackWebhookURL     string
	OnOffRampClientID      string
	OnboardingClientID     string
	ExchangeClientID       string
	APIBaseURL             string
	OnboardingBaseURL      string
	OnOffRampBaseURL       string
	EUROpsAccount          string
	EUROpsLedgerID         uint32
}

const (
	ProviderName   = "gatehub"
	AccTypeBalance = "balance"

	LedgerIDEUR   uint32 = 4482387 // Spells ghubeur on a Nokia 3320 keyboard
	EUROpsAccount        = "1854f171-eafa-4e30-bf66-7dbfe167ccfa"

	DeliveryAddressPermanentResidence = "PermanentResidence"
	DeliveryAddressTemporaryResidence = "TemporaryResidence"
	DeliveryAddressWork               = "Work"
	DeliveryAddressOther              = "Other"

	CardStatusActive           = "Active"
	CardStatusBlocked          = "Blocked"
	CardStatusTemporaryBlocked = "TemporaryBlocked"
	CardStatusReplaced         = "Replaced"
	CardStatusSoftDelete       = "SoftDelete"
	CardStatusAccountBlocked   = "AccountBlocked"
	CardStatusInCreation       = "InCreation"

	CardStatusReasonCodeClientRequestedLock           = "ClientRequestedLock"
	CardStatusReasonCodeLostCard                      = "LostCard"
	CardStatusReasonCodeStolenCard                    = "StolenCard"
	CardStatusReasonCodeIssuerRequestGeneral          = "IssuerRequestGeneral"
	CardStatusReasonCodeIssuerRequestFraud            = "IssuerRequestFraud"
	CardStatusReasonCodeIssuerRequestLegal            = "IssuerRequestLegal"
	CardStatusReasonCodeIssuerRequestIncorrectOpening = "IssuerRequestIncorrectOpening"
	CardStatusReasonCodeIssuerRequestCustomerDeceased = "IssuerRequestCustomerDeceased"
	CardStatusReasonCodeCardDamagedOrNotWorking       = "CardDamagedOrNotWorking"
	CardStatusReasonCodeUserRequest                   = "UserRequest"
	CardStatusReasonCodeProductDoesNotRenew           = "ProductDoesNotRenew"
	CardStatusReasonCodeProductChange                 = "ProductChange"
	CardStatusReasonCodeRenewed                       = "Renewed"

	CardTokenTypePin                string = "pin"
	CardTokenTypePinChange          string = "pin-change"
	CardTokenTypeCardData           string = "card-data"
	CardTokenTypeAppleProvisioning  string = "apple-provisioning"
	CardTokenTypeGoogleProvisioning string = "google-provisioning"
	CardTokenTypeDigitalizationPAI  string = "pai"

	CardLockLevelClient = "Client"
	CardLockLevelAdmin  = "Admin"

	KycAddressID = "kyc-address"

	DollarSign            = "$"
	DollarSignPlaceholder = "#" // "#" is used as a placeholder for $

	NameOnCardMaxLength int = 26

	TimeLayout = "20060102_150405" // chronological + Temporal-safe
)

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}

type CreateTransferArgs struct {
	SendingLinkedAccountID   string
	ReceivingLinkedAccountID string
	Amount                   currency.Amount
	ProviderFee              *currency.Amount
}

type User = external.User

type Card = external.Card

type CustomerDeliveryAddress = external.CustomerDeliveryAddress

type ExternalIDs struct {
	UserID           string         `db:"external_id"`
	CustomerID       sql.NullString `db:"external_customer_id"`
	CustomerSourceID sql.NullString `db:"external_customer_source_id"`
	AccountID        sql.NullString `db:"external_account_id"`
	AccountSourceID  sql.NullString `db:"external_account_source_id"`
}

func (eid ExternalIDs) IsPendingExternalCreation() bool {
	hasCustomerID := eid.CustomerID.Valid && eid.CustomerID.String != ""
	hasCustomerSourceID := eid.CustomerSourceID.Valid && eid.CustomerSourceID.String != ""
	hasAccountID := eid.AccountID.Valid && eid.AccountID.String != ""
	hasAccountSourceID := eid.AccountSourceID.Valid && eid.AccountSourceID.String != ""

	if hasCustomerSourceID && hasAccountSourceID && !hasCustomerID && !hasAccountID {
		return true
	}

	return false
}

func (eid ExternalIDs) IsCustomerCreated() bool {
	hasCustomerID := eid.CustomerID.Valid && eid.CustomerID.String != ""
	hasCustomerSourceID := eid.CustomerSourceID.Valid && eid.CustomerSourceID.String != ""
	hasAccountID := eid.AccountID.Valid && eid.AccountID.String != ""
	hasAccountSourceID := eid.AccountSourceID.Valid && eid.AccountSourceID.String != ""

	if hasCustomerID && hasCustomerSourceID && hasAccountID && hasAccountSourceID {
		return true
	}

	return false
}

type OrderCardArgs struct {
	Wallet             wallets.Wallet
	ExternalIDs        ExternalIDs
	ShouldOrderPlastic bool
	CardProductCode    string
	DeliveryAddressID  *string
	NewDeliveryAddress *external.CreateCustomerDeliveryAddressArgs
}

type GetCardTokenArgs struct {
	UserID    string
	CardID    string
	TokenType string
	PublicKey *string
}

type FreezeCardArgs struct {
	UserID     string
	CardID     string
	ReasonCode string
	Note       *string
}

type UnfreezeCardArgs struct {
	UserID string
	CardID string
	Note   *string
}

type BlockCardArgs struct {
	UserID     string
	CardID     string
	ReasonCode string
}

type CloseCardArgs struct {
	UserID     string
	CardID     string
	ReasonCode string
}

type CardTransaction struct {
	ID                  string         `db:"id"`
	CardID              string         `db:"card_id"`
	CardMaskedPan       string         `db:"card_masked_pan"`
	Type                string         `db:"type"`
	Status              *string        `db:"status"`
	Mcc                 sql.NullString `db:"mcc"`
	TransactionAmount   sql.NullString `db:"transaction_amount"`
	TransactionCurrency sql.NullString `db:"transaction_currency"`
	RawTransaction      []byte         `db:"raw_transaction"`
	CreatedAt           time.Time      `db:"created_at"`
}
