package payments

import "errors"

var (
	ErrNotFound     = errors.New("payments: Not Found")
	ErrInternal     = errors.New("payments: Internal")
	ErrInfoRequired = errors.New("payments: Information missing for payment")
)
