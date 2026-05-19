package transactions

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type TransactionType string

const (
	TransactionTypeOpenPaymentIncoming     TransactionType = "open_payments_incoming"
	TransactionTypeOpenOutgoingPayment     TransactionType = "open_payments_outgoing"
	TransactionTypeReceived                TransactionType = "received"
	TransactionTypeSent                    TransactionType = "sent"
	TransactionTypeWithdrawal              TransactionType = "withdrawal"
	TransactionTypeTransfer                TransactionType = "transfer"
	TransactionTypeDeposit                 TransactionType = "deposit"
	TransactionTypeWebMonetizationIncoming TransactionType = "web_monetization_incoming"
	TransactionTypeWebMonetizationOutgoing TransactionType = "web_monetization_outgoing"
	TransactionTypeCardTransaction         TransactionType = "card_transaction"
	TransactionTypeReturn                  TransactionType = "return"
)

type State string

const (
	StatePending   State = "Pending"
	StateCompleted State = "Completed"
	StateFailed    State = "Failed"
	StateOnHold    State = "OnHold"
)

type Provider string

const (
	ProviderOpenPayments   Provider = "open_payments"
	ProviderPaymentsEngine Provider = "payments_engine"
	ProviderXago           Provider = "xago"
)

type TransferType string

const (
	TransferTypeDebitCard            TransferType = "debit_card"
	TransferTypeCreditCard           TransferType = "credit_card"
	TransferTypeDebitBankAccount     TransferType = "debit_bank_acc"
	TransferTypeCreditBankAccount    TransferType = "credit_bank_acc"
	TransferTypeDebitWebMonetization TransferType = "debit_web_monetization"
	TransferTypeDebitBalance         TransferType = "debit_balance"
	TransferTypeCreditBalance        TransferType = "credit_balance"
)

type RefundState int16

const (
	RefundStateNone RefundState = iota
	RefundStatePending
	RefundStateCompleted
)

type CardOperation string

const (
	CardOperationWithdraw CardOperation = "Withdraw"
	CardOperationDeposit  CardOperation = "Deposit"
	CardOperationNone     CardOperation = "None"
)

func (co *CardOperation) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("unsupported CardOperation value: %v", value)
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported CardOperation value: %v", value)
	}
	switch s {
	case "0":
		*co = CardOperationWithdraw
	case "1":
		*co = CardOperationDeposit
	case "2":
		*co = CardOperationNone
	default:
		return fmt.Errorf("unsupported CardOperation value: %v", value)
	}
	return nil
}

type CardType string

const (
	CardTypePurchase                    CardType = "Purchase"
	CardTypeATMWithdrawal               CardType = "ATMWithdrawal"
	CardTypeCardVerificationInquiry     CardType = "CardVerificationInquiry"
	CardTypeCashAdvance                 CardType = "CashAdvance"
	CardTypeRefundCreditPayment         CardType = "RefundCreditPayment"
	CardTypeBalanceInquiryOnATM         CardType = "BalanceInquiryOnATM"
	CardTypePINUnblock                  CardType = "PINUnblock"
	CardTypePINChange                   CardType = "PINChange"
	CardTypePreauthorization            CardType = "Preauthorization"
	CardTypePreauthorizationIncremental CardType = "PreauthorizationIncremental"
	CardTypePreauthorizationCompletion  CardType = "PreauthorizationCompletion"
	CardTypeTransferToAccount           CardType = "TransferToAccount"
	CardTypeTransferFromAccount         CardType = "TransferFromAccount"
)

func (ct *CardType) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("unsupported CardType value: %v", value)
	}
	n, ok := value.(int64)
	if !ok {
		return fmt.Errorf("unsupported CardType value: %v", value)
	}
	switch n {
	case 0:
		*ct = CardTypePurchase
	case 1:
		*ct = CardTypeATMWithdrawal
	case 6:
		*ct = CardTypeCardVerificationInquiry
	case 17:
		*ct = CardTypeCashAdvance
	case 20:
		*ct = CardTypeRefundCreditPayment
	case 30:
		*ct = CardTypeBalanceInquiryOnATM
	case 91:
		*ct = CardTypePINUnblock
	case 92:
		*ct = CardTypePINChange
	case 101:
		*ct = CardTypePreauthorization
	case 102:
		*ct = CardTypePreauthorizationIncremental
	case 103:
		*ct = CardTypePreauthorizationCompletion
	case 107:
		*ct = CardTypeTransferToAccount
	case 108:
		*ct = CardTypeTransferFromAccount
	default:
		return fmt.Errorf("unsupported CardType value: %v", value)
	}
	return nil
}

type CreateTransactionArgs struct {
	ID                      string          `validate:"omitempty,uuid"`
	WalletID                string          `validate:"uuid"` // Fynbos wallet ID
	ForeignID               string          `validate:"omitempty"`
	ForeignType             TransactionType `validate:"required"`
	Provider                Provider        `validate:"required"`
	State                   State           `validate:"required"`
	Note                    string
	Source                  string // Usually the sending payment pointer
	Destination             string // Usually the receiving payment pointer
	Amount                  currency.Amount
	ProviderFee             *currency.Amount
	Transfers               []TransferArgs `validate:"omitempty,dive"`
	GrantID                 string
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string `validate:"omitempty,oneof=Twitter Slack wallet WalletID WalletURL ExternalWalletURL"`
	Reference               string
	Title                   string
	ExchangeRateApplied     string
	ExchangeRateReference   string
	ExchangeRateSurcharge   string
	TargetAmount            *currency.Amount
}

type UpdateTransactionArgs struct {
	WalletID        string `validate:"uuid"` // InterledgerApp wallet ID
	ForeignID       string `validate:"uuid"`
	State           State  `validate:"required"`
	Amount          currency.Amount
	ProviderFee     currency.Amount
	UpdateTransfers []TransferArgs `validate:"omitempty,dive"`
}

type UpdateForeignIDArgs struct {
	OldForeignID string `validate:"uuid"`
	NewForeignID string `validate:"uuid"`
}

type TransferArgs struct {
	LinkedAccountID string `validate:"omitempty,uuid"`
	ForeignID       string
	Type            TransferType `validate:"required"`
	Amount          currency.Amount
	State           State `validate:"required"`
}

type CardTransactionDetails struct {
	CardID        string `db:"card_id"`
	CardMaskedPan string `db:"card_masked_pan"`

	Operation      CardOperation `db:"operation"`
	Classification string        `db:"classification"`
	Type           CardType      `db:"type"`
}

// Transaction is abstract information representing either an incoming or outgoing payment, wallet top up or withdrawal.
// This is used for listing transactions for the frontend
type Transaction struct {
	ID                      string
	WalletID                string
	ForeignID               string
	Source                  string
	Destination             string
	Title                   string
	Note                    string
	Type                    TransactionType
	Timestamp               time.Time
	Provider                Provider
	State                   State
	Amount                  currency.Amount
	ProviderFee             *currency.Amount
	LinkedAccountTitle      string
	DestinationIdentity     string
	DestinationIdentityType string
	Reference               string
	RefundState             RefundState
	CardTransactionDetails  CardTransactionDetails
	ExchangeRateApplied     string
	ExchangeRateReference   string
	ExchangeRateSurcharge   string
	TargetAmount            *currency.Amount
}

// Transfer is the underlying transfers that make up a single Transactions
type Transfer struct {
	ID              string
	LinkedAccountID string
	ForeignID       string
	Type            TransferType
	Amount          currency.Amount
	ProviderFee     *currency.Amount
	State           State
	Timestamp       time.Time
}

// Filters to use when listing transactions.
type TransactionRangeFilter struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
}
