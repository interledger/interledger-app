package dynamicforms

import "errors"

var (
	ErrInternal = errors.New("dynamic forms: internal error")
	ErrNotFound = errors.New("dynamic forms: not found")
)
