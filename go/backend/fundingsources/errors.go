package fundingsources

import "errors"

var (
	ErrDuplicate       = errors.New("funding source: duplicate.")
	ErrNotFound        = errors.New("funding source: not found.")
	ErrInvalidArgument = errors.New("funding source: invalid argument.")
	ErrInternal        = errors.New("funding source: internal error.")
	ErrUnauthorized    = errors.New("funding source: unauthorized.")
)
