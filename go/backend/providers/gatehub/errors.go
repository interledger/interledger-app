package gatehub

import "errors"

var (
	ErrInternal = errors.New("gatehub: internal error")
	ErrNotFound = errors.New("gatehub: not found")
)
