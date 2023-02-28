package limits

import "errors"

var (
	ErrNotFound error = errors.New("limits: not found")
	ErrInternal error = errors.New("limits: internal error")
)
