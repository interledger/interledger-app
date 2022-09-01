package pacioli

import (
	tigerbeetleTypes "github.com/coilhq/tigerbeetle-go/pkg/types"
)

type ConfigureLedgerArgs struct {
	ID    uint32
	Name  string `validate:"required"`
	Asset string `validate:"required"`
	Scale uint8  `validate:"gt=0"`
}

type ConfigureAccountArgs struct {
	ID       string `validate:"required,uuid4"`
	LedgerID uint32 `validate:"required"`
	Code     uint16 `validate:"required"`
	Flags    AccountFlags
}

type Ledger struct {
	ID        uint32 `json:"id"` // maps to TigerBeetle's `unit` (soon to be ledger) field on an account.
	Name      string `json:"name"`
	Asset     string `json:"string"`
	Scale     uint8  `json:"scale"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Account struct {
	ID             string `db:"id"`
	LedgerID       uint32 `db:"ledger_id"`
	Flags          AccountFlags
	Code           uint16 `db:"code"`
	DebitsPending  uint64 `db:"debits_pending"`
	DebitsPosted   uint64 `db:"debits_posted"`
	CreditsPending uint64 `db:"credits_pending"`
	CreditsPosted  uint64 `db:"credits_posted"`
}

type Transfer struct {
	ID              string `db:"id"`
	PendingID       string `db:"pending_id"`
	LedgerID        uint16 `db:"ledger_id"` // this field is coming soon to a TigerBeetle near you.
	DebitAccountID  string `db:"debit_account_id"`
	CreditAccountID string `db:"credit_account_id"`
	Amount          uint64 `db:"amount"`
	Flags           TransferFlags
	Code            uint16 `db:"code"`
	Timeout         uint64
}

type CreateTransferArgs struct {
	ID              string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Flags           TransferFlags
	Code            uint16 `validate:"required"`
	Timeout         uint64
	Ledger          uint32 `validate:"required"`
}

func ToAccountFlags(in uint16) AccountFlags {
	return AccountFlags{
		Linked:                     in&(1<<0) == 1,
		DebitsMustNotExceedCredits: in&(1<<1) == 2,
		CreditsMustNotExceedDebits: in&(1<<2) == 4,
	}
}

type TransferFlags = tigerbeetleTypes.TransferFlags
type AccountFlags = tigerbeetleTypes.AccountFlags
type EventResult = tigerbeetleTypes.EventResult
type TransferResult = tigerbeetleTypes.TransferEventResult
type AccountResult = tigerbeetleTypes.AccountEventResult
type AccountResultCode = tigerbeetleTypes.CreateAccountResult
type TransferResultCode = tigerbeetleTypes.CreateTransferResult

type LedgerResult struct {
	Index uint32
	Code  LedgerResultCode
}

//go:generate stringer -type=LedgerResultCode -trimprefix=Ledger

type LedgerResultCode uint32

const (
	LedgerOK                       LedgerResultCode = 0
	LedgerExistsWithDifferentName  LedgerResultCode = 1
	LedgerExistsWithDifferentAsset LedgerResultCode = 2
	LedgerExistsWithDifferentScale LedgerResultCode = 3

	AccountOK                 AccountResultCode = 0   // TB account errors start at 1
	AccountLedgerDoesNotExist AccountResultCode = 404 // High Error number that tigerbeetle does not have defined.
)
