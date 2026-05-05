package gatehub

import "errors"

var (
	ErrInternal            = errors.New("gatehub: internal error")
	ErrNotFound            = errors.New("gatehub: not found")
	ErrInvalidWebhook      = errors.New("gatehub: invalid webhook")
	ErrInsufficientBalance = errors.New("gatehub: insufficient balance")
	ErrTimedOut            = errors.New("gatehub: timed out")
	ErrBadRequest          = errors.New("gatehub: bad request")
)
