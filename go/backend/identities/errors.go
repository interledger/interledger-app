package identities

import "errors"

var (
	ErrInternal        = errors.New("identities: internal error")
	ErrInvalidArgument = errors.New("identities: invalid argument")
	ErrNotFound        = errors.New("identities: not found")
	ErrAlreadyExists   = errors.New("identities: already exists")
)
