package payments

import "errors"

var (
	ErrNotFound               = errors.New("payments: Not Found")
	ErrInternal               = errors.New("payments: Internal")
	ErrInvalidState           = errors.New("payments: Cannot update payment confirmed or cancelled payment")
	ErrInvalidStateTransition = errors.New("payments: Invalid state transition")
	ErrRequiredActions        = errors.New("payments: Actions required")
	ErrInvalidAmount          = errors.New("payments: Invalid amount")
	ErrInvalidIdentifier      = errors.New("payments: Invalid identifier")
	ErrIdempotencyViolation   = errors.New("payments: Idempotency create violation")
	ErrInvalidWithdrawal      = errors.New("payments: Invalid withdrawal accounts")
	ErrIncompatibleAccounts   = errors.New("payments: Incompatible sender receiver accounts")
	ErrInsufficientFunds      = errors.New("payments: Insufficient funds")
)
