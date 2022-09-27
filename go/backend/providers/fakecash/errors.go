package fakecash

import "errors"

var (
	ErrInternal = errors.New("fake cash provider: internal error.")
	ErrNotFound = errors.New("fake cash provider: not found.")
)
