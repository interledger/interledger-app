package pti

import (
	"context"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti/external"
)

// TODO: asked for 12 may 2025
const (
	ProviderName   = "pti"
	AccTypeBalance = "balance"
	TypeCard       = "card"
	TypeBank       = "bank_account"

	// TODO: Ask?
	ScenarioTransfer   = "ilf_transfer"
	ScenarioDeposit    = "ilf_deposit"
	ScenarioWithdrawal = "ilf_withdrawal"

	LedgerIDUSD   uint32 = 784873 // Spells ptiusd on a Nokia 3320 keyboard
	USDOpsAccount        = "fb4713ba-94c5-4a56-a5bf-82b551e9bd40"
)

type TransactionFeedback string

const (
	TransactionFeedbackAccepted    TransactionFeedback = "ACCEPTED"
	TransactionFeedbackSettled     TransactionFeedback = "SETTLED"
	TransactionFeedbackCancelled   TransactionFeedback = "CANCELLED"
	TransactionFeedbackRejected    TransactionFeedback = "REJECTED"
	TransactionFeedbackRefunded    TransactionFeedback = "REFUNDED"
	TransactionFeedbackChargedBack TransactionFeedback = "CHARGED_BACK"
	TransactionFeedbackError       TransactionFeedback = "ERROR"
)

type User struct {
	ID               string `db:"id"`
	ExternalID       string `db:"external_id"`
	WalletID         string `db:"wallet_id"`
	Status           string `db:"status"`
	AssessmentStatus string `db:"assessment_status"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}

type Wallet struct {
	ID        string            `db:"id"`
	UserID    string            `db:"external_user_id"`
	Reference string            `db:"reference"`
	Currency  currency.Currency `db:"currency"`
	CreatedAt time.Time         `db:"created_at"`
}

type Await func(ctx context.Context, result interface{}) error

type CreateWalletArgs struct {
	WalletID string
	Currency currency.Currency
	Nickname string
	Title    string
}

type CreateExternalWalletArgs struct {
	ID       string
	UserID   string
	Currency currency.Currency
}

type TransactionArgs struct {
	PaymentID       string
	WalletID        string
	Amount          currency.Amount
	LinkedAccountID string
	Note            string
}

type TransactionStatusArgs struct {
	PaymentID     string
	TransactionID string
	Status        TransactionFeedback
	Amount        currency.Amount
}

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}

type TokenArgs = external.TokenArgs
type TokenResponse = external.TokenResponse
type EncryptedCreditCardPaymentInformation = external.EncryptedCreditCardPaymentInformation
type WidgetDetails = struct {
	ScenarioID        string
	RequestID         string
	UserID            string
	GenerateTokenPath string
	ClientID          string
	SdkUrl            string
	FormsUrl          string
	SessionID         string
}

type CreateBankAccountArgs struct {
	WalletID                string
	AccountNumber           string
	AccountType             string
	RoutingNumber           string
	RoutingNumberCheckDigit string
	Bank                    string
}

type TransactionStatusPayload struct {
	ResourceType       string       `json:"resourceType"`
	RequestID          string       `json:"requestId"`
	ClientID           string       `json:"clientId"`
	UserID             string       `json:"userId"`
	Status             string       `json:"status"`
	Date               FlexibleTime `json:"date"`
	Amount             float64      `json:"amount"`
	Fees               string       `json:"fees"`
	Currency           string       `json:"currency"`
	TransactionType    string       `json:"transactionType"`
	PaymentMethod      string       `json:"paymentMethod"`
	TransactionGroupID string       `json:"transactionGroupId"`
	// AdditionalInfos    struct {
	// 	CreditCardLast4Digits        string `json:"CreditCardLast4Digits"`
	// 	PaymentProviderTransactionID string `json:"PaymentProviderTransactionId"`
	// } `json:"additionalInfos"`
	// PaymentStatusDetail struct {
	// 	ProviderResponseCode     string `json:"providerResponseCode"`
	// 	ProviderResponseCategory string `json:"providerResponseCategory"`
	// } `json:"paymentStatusDetail"`
	TransactionTotal struct {
		Total struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"total"`
		Fee struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"fee"`
		Subtotal struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"subtotal"`
	} `json:"transactionTotal,omitempty"`
	Total struct {
		SubTotal struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"subTotal"`
		Fee struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"fee"`
		Total struct {
			Amount   int    `json:"amount"`
			Currency string `json:"currency"`
		} `json:"total"`
	} `json:"total,omitempty"`
}

type FlexibleTime struct {
	time.Time
}

func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	var t time.Time
	var err error

	// Try full RFC3339 first
	t, err = time.Parse(time.RFC3339, s)
	if err != nil {
		// Fallback for "2025-10-18T19:07Z"
		t, err = time.Parse("2006-01-02T15:04Z07:00", s)
	}
	if err != nil {
		return err
	}

	ft.Time = t
	return nil
}
