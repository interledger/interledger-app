package linkedaccounts

import "errors"

var (
	ErrDuplicate       = errors.New("linked account: duplicate.")
	ErrNotFound        = errors.New("linked account: not found.")
	ErrInvalidArgument = errors.New("linked account: invalid argument.")
	ErrInternal        = errors.New("linked account: internal error.")
	ErrUnauthorized    = errors.New("linked account: unauthorized.")
)
