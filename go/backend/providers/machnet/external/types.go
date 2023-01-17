package external

import (
	"encoding/json"

	"gitlab.com/fynbos/backend/currency"
)

const (
	StatusVerified   = "VERIFIED"
	StatusUnverified = "UNVERIFIED"
	StatusFailed     = "FAILED"
	StatusPending    = "PENDING"

	TypeSendUser    = "SEND"
	TypeReceiveUser = "RECEIVE"

	TypeBankDeposit = "BANK_DEPOSIT"

	TypeCard   = "CARD"
	TypeBank   = "BANK"
	TypeWallet = "WALLET"

	DeliveryStatusNone      = "NONE"
	DeliveryStatusRequested = "DELIVERY_REQUESTED"
)

type FundingSourceType string

const (
	FundingSourceTypeCard        FundingSourceType = "CARD"
	FundingSourceTypeBankAccount FundingSourceType = "BANK_ACCOUNT"
)

type PayoutMethod string

const (
	PayoutMethodBankDeposit  PayoutMethod = "BANK_DEPOSIT"
	PayoutMethodCashPickup   PayoutMethod = "CASH_PICKUP"
	PayoutMethodWallet       PayoutMethod = "WALLET"
	PayoutMethodHomeDelivery PayoutMethod = "HOME_DELIVERY"
)

type CalculationMode string

const (
	CalculationModeSenderAmount   CalculationMode = "SENDER_AMOUNT"
	CalculationModeReceiverAmount CalculationMode = "RECEIVER_AMOUNT"
)

type Purpose string

const (
	PurposePersonalTransfer Purpose = "PERSONAL_TRANSFER"
)

type AccountType string

const (
	AccountTypeCheque  AccountType = "CHECKING"
	AccountTypeSavings AccountType = "SAVINGS"
)

type CreateTransactionArgs struct {
	FromUserID        string            `json:"-"`
	FromFundID        string            `json:"from_fund_id,omitempty"`
	FundingSourceType FundingSourceType `json:"funding_source_type,omitempty"`
	FromAmount        float64           `json:"from_amount,omitempty"`
	FromCurrency      string            `json:"from_currency,omitempty"`
	ToCurrency        string            `json:"to_currency,omitempty"`
	ExchangeRate      float64           `json:"exchange_rate,omitempty"`
	FeeAmount         float64           `json:"fee_amount"`
	Purpose           Purpose           `json:"purpose"`
	IPAddress         string            `json:"ip_address,omitempty"`
	To                TransactionTo     `json:"to"`
}

type User struct {
	ID                string             `json:"id,omitempty"`
	FirstName         string             `json:"first_name,omitempty"`
	MiddleName        string             `json:"middle_name,omitempty"`
	LastName          string             `json:"last_name,omitempty"`
	Email             string             `json:"email,omitempty"`
	Gender            string             `json:"gender,omitempty"`
	DateOfBirth       string             `json:"date_of_birth,omitempty"`
	AddressLine1      string             `json:"address_line1,omitempty"`
	AddressLine2      string             `json:"address_line2,omitempty"`
	MobilePhone       string             `json:"mobile_phone,omitempty"`
	City              string             `json:"city,omitempty"`
	Zipcode           string             `json:"zipcode,omitempty"`
	State             string             `json:"state,omitempty"`
	Country           string             `json:"country,omitempty"`
	IPAddress         string             `json:"ip_address,omitempty"`
	Type              string             `json:"type,omitempty"`
	PhysicalDocuments []PhysicalDocument `json:"physical_documents,omitempty"`
	Status            string             `json:"status,omitempty"`
	SendUserID        string             `json:"send_user_id,omitempty"`
	Business          bool               `json:"business"`
	BusinessType      string             `json:"business_type,omitempty"`
}

type PhysicalDocument struct {
	DocumentValue     string `json:"document_value,omitempty"`
	DocumentValueBack string `json:"document_value_back,omitempty"`
	DocumentType      string `json:"document_type,omitempty"`
	DocumentCountry   string `json:"country,omitempty"`
	DocumentState     string `json:"state,omitempty"`
}

type CipInfo struct {
	Gender       string `json:"GENDER,omitempty"`
	Country      string `json:"COUNTRY,omitempty"`
	State        string `json:"STATE,omitempty"`
	LastName     string `json:"LAST_NAME,omitempty"`
	Occupation   string `json:"OCCUPATION,omitempty"`
	City         string `json:"CITY,omitempty"`
	ZipCode      string `json:"ZIP_CODE,omitempty"`
	DateOfBirth  string `json:"DATE_OF_BIRTH,omitempty"`
	IDDoc        string `json:"ID_DOC,omitempty"`
	PhoneNumber  string `json:"PHONE_NUMBER,omitempty"`
	Email        string `json:"EMAIL,omitempty"`
	AddressLine1 string `json:"ADDRESS_LINE1,omitempty"`
	AddressLine2 string `json:"ADDRESS_LINE2,omitempty"`
	MiddleName   string `json:"MIDDLE_NAME,omitempty"`
	FirstName    string `json:"FIRST_NAME,omitempty"`
	IPAddress    string `json:"IP_ADDRESS,omitempty"`
}

type VerificationStatus struct {
	CipInfo   CipInfo `json:"cip_info,omitempty"`
	KycStatus string  `json:"kyc_status,omitempty"`
	UserID    string  `json:"user_id,omitempty"`
}

type InitiateKycResponse struct {
	Success bool   `json:"success,omitempty"`
	Status  string `json:"status,omitempty"`
}

type WidgetTokenResponse struct {
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

type FundingSource struct {
	ID                 string `json:"id,omitempty"`
	UserID             string `json:"user_id,omitempty"`
	FundingsourceName  string `json:"funding_source_name,omitempty"`
	FundingsourceType  string `json:"funding_source_type,omitempty"`
	AccountNumber      string `json:"account_number,omitempty"`
	InstitutionName    string `json:"institution_name,omitempty"`
	VerificationStatus string `json:"verification_status,omitempty"`
}

type Transaction struct {
	ID                string            `json:"id,omitempty"`
	UserID            string            `json:"user_id,omitempty"`
	FromAmount        float64           `json:"from_amount,omitempty"`
	FromCurrency      string            `json:"from_currency,omitempty"`
	ToAmount          float64           `json:"to_amount,omitempty"`
	ToCurrency        string            `json:"to_currency,omitempty"`
	FeeAmount         float64           `json:"fee_amount,omitempty"`
	ExchangeRate      float64           `json:"exchange_rate,omitempty"`
	FromFundID        string            `json:"from_fund_id,omitempty"`
	FundingsourceType FundingSourceType `json:"funding_source_type,omitempty"`
	DeliveryStatus    string            `json:"delivery_status,omitempty"`
	Status            string            `json:"status,omitempty"`
	IPAddress         string            `json:"ip_address,omitempty"`
	To                TransactionTo     `json:"to"`
}

const (
	TransactionInitiated  = "INITIATED"
	TransactionPending    = "PENDING"
	TransactionProcessing = "PROCESSING"
	TransactionProcessed  = "PROCESSED"
	TransactionCancelled  = "CANCELED"
	TransactionFailed     = "FAILED"
	TransactionHold       = "HOLD"
	TransactionRefunded   = "REFUNDED"
	TransactionReturned   = "RETURNED"

	TransactionPendingEvent    = "transaction_pending"
	TransactionProcessingEvent = "transaction_processing"
	TransactionProcessedEvent  = "transaction_processed"
	TransactionCancelledEvent  = "transaction_canceled"
	TransactionFailedEvent     = "transaction_failed"
	TransactionHoldEvent       = "transaction_onhold"
	TransactionReturnedEvent   = "transaction_returned"

	TransactionDeliveryNone        = "NONE"
	TransactionDeliveryHold        = "HOLD"
	TransactionDeliveryPending     = "PENDING"
	TransactionDeliveryRequested   = "DELIVERY_REQUESTED"
	TransactionDeliveryDelivered   = "DELIVERED"
	TransactionDeliveryFailed      = "DELIVERY_FAILED"
	TransactionDeliveryAuthorized  = "DELIVERY_AUTHORIZED"
	TransactionDeliveryPayoutReady = "DELIVERY_PAYOUT_READY"

	TransactionDeliveryHoldEvent        = "transaction_delivery_onhold"
	TransactionDeliveryPendingEvent     = "transaction_delivery_pending"
	TransactionDeliveryRequestedEvent   = "transaction_delivery_requested"
	TransactionDeliveredEvent           = "transaction_delivered"
	TransactionDeliveryFailedEvent      = "transaction_delivery_failed"
	TransactionDeliveryAuthorizedEvent  = "transaction_delivery_authorized "
	TransactionDeliveryPayoutReadyEvent = "transaction_delivery_payout_ready"
)

type TransactionTo struct {
	AddressLine1    string          `json:"address_line1,omitempty"`
	CalculationMode CalculationMode `json:"calculation_mode,omitempty"`
	Email           string          `json:"email,omitempty"`
	FirstName       string          `json:"first_name,omitempty"`
	FundID          string          `json:"fund_id,omitempty"`
	ID              string          `json:"id,omitempty"`
	LastName        string          `json:"last_name,omitempty"`
	MobilePhone     string          `json:"mobile_phone,omitempty"`
	PayoutMethod    PayoutMethod    `json:"payout_method,omitempty"`
}

type ReceiveUserBankAccount struct {
	ID            string      `json:"id,omitempty"`
	UserID        string      `json:"user_id,omitempty"`
	AccountNumber string      `json:"account_number,omitempty"`
	AccountType   AccountType `json:"account_type"`
	BankID        int         `json:"bank_id"`
	BranchID      int         `json:"branch_id"`
	PayoutMethod  string      `json:"payout_method"`
}

type DeliveryRequest struct {
	Status        string `json:"status"`
	TransactionID string `json:"-"`
	UserID        string `json:"-"`
}

type (
	Event struct {
		ID             string          `json:"id" db:"db"`
		EventName      string          `json:"event_name" db:"event_name"`
		ResourceID     string          `json:"resource_id" db:"resource_id"`
		UserID         string          `json:"user_id" db:"user_id"`
		Payload        json.RawMessage `json:"payload" db:"payload"` // TODO: find out what payload looks like
		SubscriptionID string          `json:"subscription_id" db:"subscription_id"`
		Timestamp      string          `json:"timestamp"`
	}
)

const (
	UserCardAdded              = "user_card_added"
	UserCardRemoved            = "user_card_removed"
	UserBankAdded              = "user_bank_added"
	UserBankRemoved            = "user_bank_removed"
	UserBankVerificationFailed = "user_bank_verification_failed"

	UserKYCInProgressEvent    = "user_kyc_in_progress"
	UserKYCVerifiedEvent      = "user_kyc_verified"
	UserKYCRetryEvent         = "user_kyc_retry"
	UserKYCSuspendedEvent     = "user_kyc_suspended"
	UserKYCReviewPendingEvent = "user_kyc_review_pending"

	UserKYCUnverified    = "UNVERIFIED"
	UserKYCInProgress    = "IN_PROGRESS"
	UserKYCVerified      = "VERIFIED"
	UserKYCRetry         = "RETRY"
	UserKYCSuspended     = "SUSPENDED"
	UserKYCReviewPending = "REVIEW_PENDING"
)

type Branch struct {
	ID   uint32 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Bank struct {
	ID                        uint32   `json:"id,omitempty"`
	Name                      string   `json:"name,omitempty"`
	Branches                  []Branch `json:"branches,omitempty"`
	Country                   string   `json:"country,omitempty"`
	ReceivingCurrency         []string `json:"receiving_currency,omitempty"`
	TransactionSupportedTypes []string `json:"txn_supported_types,omitempty"`
}

type (
	Wallet struct {
		ID                 string        `json:"id,omitempty"`
		UserID             string        `json:"user_id,omitempty"`
		NickName           string        `json:"nick_name,omitempty"`
		FundingSourceType  string        `json:"funding_source_type,omitempty"`
		VerificationStatus string        `json:"verification_status,omitempty"`
		Balance            WalletBalance `json:"balance"`
	}

	WalletBalance struct {
		AvailableBalance float64 `json:"available_balance"`
		Balance          float64 `json:"balance"`
	}

	WalletTransferArgs struct {
		SendUserID string
		SendFundID string
		RecvUserID string
		RecvFundID string
		Amount     currency.Amount
		Currency   string
		IPAddress  string
	}

	WalletTransfer struct {
		ID         string        `json:"id,omitempty"`
		UserID     string        `json:"user_id,omitempty"`
		Amount     float64       `json:"amount,omitempty"`
		Currency   string        `json:"currency,omitempty"`
		FeeAmount  float64       `json:"fee_amount,omitempty"`
		FromFundID string        `json:"from_fund_id,omitempty"`
		Status     string        `json:"status,omitempty"`
		IPAddress  string        `json:"ip_address,omitempty"`
		To         TransactionTo `json:"to"`
		Type       string        `json:"type"`
	}

	FundWalletArgs struct {
		UserID       string
		SourceFundID string
		WalletID     string
		Amount       float64
		Currency     string
		IPAddress    string
	}

	FundWalletResponse struct {
		ID           string  `json:"id"`
		UserID       string  `json:"user_id"`
		SourceFundID string  `json:"from_fund_id"`
		Status       string  `json:"status"`
		Amount       float64 `json:"amount"`
		Currency     string  `json:"currency"`
		IPAddress    string  `json:"ip_address"`
		Type         string  `json:"type"`
	}
)

type (
	WalletWithdrawal struct {
		ID           string             `json:"id"`
		UserID       string             `json:"user_id"`
		SourceFundID string             `json:"from_fund_id"`
		Status       string             `json:"status"`
		Amount       float64            `json:"amount"`
		FeeAmount    float64            `json:"fee_amount"`
		Currency     string             `json:"currency"`
		IPAddress    string             `json:"ip_address"`
		Type         string             `json:"type"`
		To           WalletWithdrawalTo `json:"to"`
	}

	WalletWithdrawalTo struct {
		UserID string `json:"id"`
		FundID string `json:"fund_id"`
	}

	WithdrawFromUserWalletArgs struct {
		UserID    string
		ToFundID  string
		WalletID  string
		Amount    float64
		FeeAmount float64
		Currency  string
		IPAddress string
	}
)
