package twitter

import "errors"

var (
	ErrInternal = errors.New("twitter: internal error")
	ErrNotFound = errors.New("twitter: not found")
)
