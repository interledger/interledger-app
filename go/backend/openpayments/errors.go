package openpayments

import "errors"

var (
	ErrPaymentPointerExists   = errors.New("payment pointer exists already")
	ErrInvalidPointerURL      = errors.New("invalid URL for format for payment pointer")
	ErrPaymentPointerNotFound = errors.New("payment pointer not found")
	ErrInternal               = errors.New("payment pointer internal error")
	ErrInvalidArgument        = errors.New("invalid argument")
)
