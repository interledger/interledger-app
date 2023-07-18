package transactions

import "errors"

var (
	ErrInvalidArgument = errors.New("transactions: invalid argument")
	ErrInternal        = errors.New("transactions: internal error")
	ErrNotFound        = errors.New("transactions: not found")
)
