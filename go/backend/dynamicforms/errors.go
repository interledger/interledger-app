package dynamicforms

import "errors"

var (
	ErrInternal = errors.New("identities: internal error")
	ErrNotFound = errors.New("identities: not found")
)
