package waitlist

import "errors"

var (
	ErrInvalidEmail   = errors.New("invalid email address format")
	ErrInvalidCountry = errors.New("invalid country code")
	ErrInvalidName    = errors.New("invalid name")
	ErrInternal       = errors.New("internal error")
)
