package astra

import "errors"

var (
	ErrInternal = errors.New("astra: internal error")
	ErrNotFound = errors.New("astra: not found")
)
