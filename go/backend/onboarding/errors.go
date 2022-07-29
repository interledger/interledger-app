package onboarding

import "errors"

var (
	ErrInternal        = errors.New("onboarding: internal error.")
	ErrInvalidArgument = errors.New("onboarding: invalid argument.")
	ErrNotFound        = errors.New("onboarding: not found.")
)
