package openpayments

import "errors"

var (
	ErrNotFound                 = errors.New("open payments: not found")
	ErrPaymentPointerExists     = errors.New("payment pointer: exists already")
	ErrPaymentPointerCannotRecv = errors.New("payment pointer: not enabled for receiving")
	ErrPaymentPointerCannotSend = errors.New("payment pointer: not enabled for sending")
	ErrInternal                 = errors.New("open payments: internal error")
	ErrInvalidArgument          = errors.New("open payments: invalid argument")
	ErrInsufficientBalance      = errors.New("open payments: insufficient balance")
)
