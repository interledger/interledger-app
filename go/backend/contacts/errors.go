package contacts

import "errors"

var (
	ErrInternal        = errors.New("internal error")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
)
