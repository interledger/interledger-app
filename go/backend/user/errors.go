package user

import "errors"

var (
	ErrInternal     = errors.New("user: internal error")
	ErrNoUserFound  = errors.New("no user found")
	ErrAAL1Required = errors.New("aal1 authentication required")
	ErrAAL2Required = errors.New("aal2 authentication required")
)
