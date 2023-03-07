package identities

import "errors"

var (
	ErrInternal        = errors.New("identities: internal error")
	ErrInvalidArgument = errors.New("identities: invalid argument")
)
