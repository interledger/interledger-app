package waitlist

import "errors"

var (
	ErrInvalidEmail   = errors.New("invalid email address format")
	ErrInvalidCountry = errors.New("invalid country code")
	ErrInvalidName    = errors.New("invalid name")
	ErrNotFound       = errors.New("not found")
	ErrInternal       = errors.New("internal error")
)
