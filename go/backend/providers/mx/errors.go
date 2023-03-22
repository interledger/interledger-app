package mx

import "errors"

var (
	ErrInternal            = errors.New("mx: internal error.")
	ErrNotFound            = errors.New("mx: not found.")
	ErrInvalidAccountOwner = errors.New("mx: invalid account owner.")
)
