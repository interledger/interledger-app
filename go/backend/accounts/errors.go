package accounts

import "errors"

var (
	ErrInternal        = errors.New("accounts ops: internal error.")
	ErrDuplicate       = errors.New("accounts ops: duplicate.")
	ErrInvalidArgument = errors.New("accounts ops: invalid argument.")
	ErrNotFound        = errors.New("accounts ops: not found.")
)
