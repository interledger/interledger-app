package pti

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti/external"
)

const (
	ProviderName   = "pti"
	AccTypeBalance = "balance"
	TypeCard       = "card"
	TypeBank       = "bank_account"

	ScenarioTransfer   = "fynbos_transfer"
	ScenarioDeposit    = "fynbos_deposit"
	ScenarioWithdrawal = "fynbos_withdrawal"

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
}

type CreateBankAccountArgs struct {
	WalletID                string
	AccountNumber           string
	AccountType             string
	RoutingNumber           string
	RoutingNumberCheckDigit string
	Bank                    string
}
