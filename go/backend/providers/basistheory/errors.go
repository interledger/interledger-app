package basistheory

import "errors"

var (
	ErrInternal            = errors.New("basistheory: internal error")
	ErrInvalidArgument     = errors.New("basistheory: invalid argument")
	ErrUserHasExistingCard = errors.New("basistheory: user has an existing card")
)
