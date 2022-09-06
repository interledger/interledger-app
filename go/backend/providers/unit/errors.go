package unit

import "errors"

var (
	ErrInternal        = errors.New("unit: internal error")
	ErrUnauthorized    = errors.New("unit: unauthorized webhook request")
	ErrInvalidArgument = errors.New("unit: invalid argument")
	ErrClient          = errors.New("unit: client error")
	ErrServer          = errors.New("unit: server error")
	ErrTimeout         = errors.New("unit: timeout error")
	ErrRateLimit       = errors.New("unit: rate limit error")
	ErrNotFound        = errors.New("unit: not found")
	ErrDuplicateEvent  = errors.New("unit webhook: duplicate event.") // event already stored in database.
)
