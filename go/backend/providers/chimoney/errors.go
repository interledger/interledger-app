package chimoney

import "errors"

var (
	ErrInternal = errors.New("chimoney: internal error")
	ErrNotFound = errors.New("chimoney: not found")
)
