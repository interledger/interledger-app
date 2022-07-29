package country

import "errors"

var (
	ErrNotFound        = errors.New("country: not found.")
	ErrInvalidArgument = errors.New("country: invalid argument.")
	ErrInternal        = errors.New("country: internal error.")
)
