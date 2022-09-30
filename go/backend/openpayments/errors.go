package openpayments

import "errors"

var (
	ErrPaymentPointerExists   = errors.New("payment pointer: exists already")
	ErrInvalidPointerURL      = errors.New("payment pointer: invalid URL format")
	ErrInvalidPointerPath     = errors.New("payment pointer: invalid user  path")
	ErrPaymentPointerNotFound = errors.New("payment pointer: not found")
	ErrInternal               = errors.New("payment pointer: internal error")
	ErrInvalidArgument        = errors.New("payment pointer: invalid argument")
)
