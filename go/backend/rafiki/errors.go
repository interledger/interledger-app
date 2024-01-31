package rafiki

import "errors"

var (
	ErrInternal = errors.New("rafiki: internal")
	ErrNotFound = errors.New("rafiki: not found")
)
