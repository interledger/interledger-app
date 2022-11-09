package email

import "errors"

var (
	ErrInvalidTemplate = errors.New("email: invalid template")
	ErrInternal        = errors.New("email: internal error")
)
