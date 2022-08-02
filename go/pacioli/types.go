package pacioli

import tigerbeetleTypes "github.com/coilhq/tigerbeetle-go/pkg/types"

type ConfigureLedgerArgs struct {
	ID    uint32
	Name  string `validate:"required"`
	Asset string `validate:"required"`
	Scale uint8  `validate:"gt=0"`
}

type ConfigureAccountArgs struct {
	ID       string `validate:"required,uuid4"`
	LedgerID uint32
	Code     uint16
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
	ID             string
	LedgerID       uint32
	Flags          AccountFlags
	Code           uint16
	DebitsPending  uint64
	DebitsPosted   uint64
	CreditsPending uint64
	CreditsPosted  uint64
}

type Transfer struct {
	ID              string
	PendingID       string
	LedgerID        uint16 // this field is coming soon to a TigerBeetle near you.
	DebitAccountID  string
	CreditAccountID string
	Amount          uint64
	Flags           TransferFlags
	Code            uint16
	Timeout         uint64
}

type CreateTransferArgs struct {
	ID              string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Flags           TransferFlags
	Code            uint16
	Timeout         uint64
	Ledger          uint32
}

type TransferFlags = tigerbeetleTypes.TransferFlags
type AccountFlags = tigerbeetleTypes.AccountFlags
type EventResult = tigerbeetleTypes.EventResult
type TransferResult = tigerbeetleTypes.TransferEventResult
type AccountResult = tigerbeetleTypes.AccountEventResult
type AccountResultCode = tigerbeetleTypes.CreateAccountResult
type TransferResultCode = tigerbeetleTypes.CreateAccountResult

const (
	LEDGER_OK                          uint8 = 0
	LEDGER_EXISTS_WITH_DIFFERENT_NAME  uint8 = 1
	LEDGER_EXISTS_WITH_DIFFERENT_ASSET uint8 = 2
	LEDGER_EXISTS_WITH_DIFFERENT_SCALE uint8 = 3

	ACCOUNT_LEDGER_DOES_NOT_EXIST uint8 = 0 // TB account errors start at 1
)
