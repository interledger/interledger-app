package slack

import "errors"

var (
	ErrInternal = errors.New("slack: internal error")
	ErrNotFound = errors.New("slack: not found")
)
