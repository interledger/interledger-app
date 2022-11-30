package openpayments

import "errors"

var (
	ErrNotFound                 = errors.New("open payments: not found")
	ErrPaymentPointerExists     = errors.New("payment pointer: exists already")
	ErrInvalidPointerURL        = errors.New("payment pointer: invalid URL format")
	ErrInvalidPointerPath       = errors.New("payment pointer: invalid user  path")
	ErrPaymentPointerNotFound   = errors.New("payment pointer: not found")
	ErrPaymentPointerCannotRecv = errors.New("payment pointer: not enabled for receiving")
	ErrPaymentPointerCannotSend = errors.New("payment pointer: not enabled for sending")
	ErrInternal                 = errors.New("open payments: internal error")
	ErrInvalidArgument          = errors.New("open payments: invalid argument")
	ErrInsufficientBalance       = errors.New("open payments: insufficient balance")
)
