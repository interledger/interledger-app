package machnet

import "errors"

var (
	ErrInternal = errors.New("machnet: internal error.")
	ErrNotFound = errors.New("machnet: not found.")
)
