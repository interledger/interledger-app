package keys

import "errors"

var (
	ErrInternal  = errors.New("keys: internal error")
	ErrNotFound  = errors.New("keys: not found")
	ErrDuplicate = errors.New("keys: duplicate key")
)
