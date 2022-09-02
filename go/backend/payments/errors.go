package payments

import "errors"

var (
	ErrInvalidArgument     = errors.New("payments service: invalid argument.")
	ErrInternal            = errors.New("payments service: internal error.")
	ErrUnverifiedAccount   = errors.New("payments service: unverified account.")
	ErrInsufficientBalance = errors.New("payments service: insufficient balance.")
	ErrUnauthorized        = errors.New("payments service: unauthorized.")
)
