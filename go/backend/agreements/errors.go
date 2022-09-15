package agreements

import "errors"

var (
	ErrInternal        = errors.New("agreements: internal error")
	ErrInvalidArgument = errors.New("agreements: invalid argument")
	ErrNotFound        = errors.New("agreements: not found")
)
