package rafiki

import "errors"

var (
	ErrInternal                  = errors.New("rafiki: internal")
	ErrNotFound                  = errors.New("rafiki: not found")
	ErrCrossCurrencyNotSupported = errors.New("rafiki: cross currency not supported")
	ErrCurrencyNotSupported      = errors.New("rafiki: currency not supported")
	ErrTimedOut                  = errors.New("rafiki: timed out")
)
