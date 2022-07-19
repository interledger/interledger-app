package accounts

import "errors"

var (
	ErrInternal        = errors.New("accounts service: internal error.")
	ErrDuplicate       = errors.New("accounts service: duplicate.")
	ErrInvalidArgument = errors.New("accounts service: invalid argument.")
	ErrNotFound        = errors.New("accounts service: not found.")
)
