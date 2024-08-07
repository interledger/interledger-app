package xago

import "errors"

var (
	ErrInternal            = errors.New("xago: internal error")
	ErrNotFound            = errors.New("xago: not found")
	ErrInsufficientBalance = errors.New("xago: insufficient balance")
	ErrTimedOut            = errors.New("xago: timed out")
)
