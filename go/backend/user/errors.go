package user

import "errors"

var (
	ErrInternal    = errors.New("user: internal error")
	ErrNoUserFound = errors.New("no user found")
)
