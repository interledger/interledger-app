package gmt

import "errors"

var (
	ErrNotFound = errors.New("gmt: not found.")
	ErrInternal = errors.New("gmt: internal error.")
)
