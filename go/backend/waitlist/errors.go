package waitlist

import "errors"

var (
	ErrInvalidEmail   = errors.New("invalid email address format")
	ErrInvalidCountry = errors.New("invalid country code")
	ErrInternal       = errors.New("internal error")
)
