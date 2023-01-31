package authorisation

import "errors"

var (
	ErrInternal        = errors.New("authorisation: internal error")
	ErrNotFound        = errors.New("authorisation: not found")
	ErrInvalidArgument = errors.New("authorisation: invalid argument")
)
