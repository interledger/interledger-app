package identity

import "errors"

var (
	ErrInternal        = errors.New("identity: internal error.")
	ErrInvalidArgument = errors.New("identity: invalid argument.")
	ErrNotFound        = errors.New("identity: not found.")
	ErrDuplicate       = errors.New("identity: duplicate.")
)
