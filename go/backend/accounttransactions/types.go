package account_transactions

// This represents the transaction that a user has performed on their account.
// e.g. deposit, withdrawal, send etc. These HAVE to be backed by one or more ledger transfers
// that are stored in TigerBeetle.
type AccountTransaction struct {
	ID          string
	Type        string
	AccountID   string `db:"account_id"`
	Description string
	State       State

	// This is the net position change on the account from the user's point of view.
	NetAmount   int64    `db:"net_amount"`
	TransferIDs []string `db:"transfer_ids"` // The ids of the transfers that are stored in TigerBeetle.
	CreatedAt   string   `db:"created_at"`
	UpdatedAt   string   `db:"updated_at"`
}

type LedgerTransferFlags struct { // duplicate of Pacioli.TransferFlags
	Linked         bool
	TwoPhaseCommit bool // TODO: update to latest TigerBeetle interface
	Condition      bool
}

// Arguments to create a transfer in TigerBeetle
type CreateLedgerTransferArgs struct {
	LedgerID        uint32 `validate:"required"`
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required"`
	Code            uint16
	Flags           LedgerTransferFlags
}

type CreateTransactionArgs struct {
	AccountID   string `validate:"required,uuid4"`
	Description string
	Type        string `validate:"oneof=deposit withdrawal outgoingPayment"`

	// a uint64 as you can't have a negative deposit/withdrawal etc.
	NetAmount uint64 `validate:"gt=0"`

	// We assume an account transaction has to backed by at least one ledger transfer
	LedgerTransfers []CreateLedgerTransferArgs `validate:"dive,gt=0"`
}

type CreatePendingTransactionArgs struct {
	AccountID   string `validate:"required,uuid4"`
	Description string
	Type        string `validate:"oneof=deposit withdrawal outgoingPayment"`

	// a uint64 as you can't have a negative deposit/withdrawal etc.
	NetAmount uint64 `validate:"gt=0"`

	// We assume an account transaction has to backed by at least one ledger transfer
	LedgerTransfers []CreateLedgerTransferArgs `validate:"dive,gt=0"`
}

type GetByAccountArgs struct {
	AccountID string `validate:"required,uuid4"`
	Limit     uint32 `validate:"gt=1"`
	OrderBy   string `validate:"oneof=ASC DESC"`
}

type PaginationArgs struct {
	AccountID string `validate:"required,uuid4"`
	After     string `validate:"omitempty,uuid4"` // forward pagination cursor
	First     uint32 `validate:"lt=1000"`         // forward pagination limit. max 1000
}

type State string

const (
	Pending = State("PENDING")
	Posted  = State("POSTED")
	Voided  = State("VOIDED")
)

func (s State) String() string {
	return string(s)
}

func (s State) IsValid() bool {
	switch s {
	case Pending, Posted, Voided:
		return true
	}
	return false
}
