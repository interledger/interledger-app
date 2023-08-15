package payments

import "errors"

var (
	ErrNotFound               = errors.New("payments: Not Found")
	ErrInternal               = errors.New("payments: Internal")
	ErrInfoRequired           = errors.New("payments: Information missing for payment")
	ErrInvalidStateTransition = errors.New("payments: Invalid state transition")
	ErrRequiredActions        = errors.New("payments: Actions required")
)
