package account_transactions

import "errors"

var (
	ErrInternal              = errors.New("account transactions: internal error")
	ErrInvalidArgument       = errors.New("account transactions: invalid argument")
	ErrNotFound              = errors.New("account transactions: not found")
	ErrDuplicate             = errors.New("account transactions: duplicate")
	ErrInvalidLedgerTransfer = errors.New("account transactions: ledger transfer failed")
	ErrExceedsDebits         = errors.New("account transactions: exceeds debits")
	ErrExceedsCredits        = errors.New("account transactions: exceeds credits")
)
