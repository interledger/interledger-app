package auth

import (
	"errors"
)

var (
	ErrMissingToken       = errors.New("missing authorization token")
	ErrInvalidFormat      = errors.New("invalid authorization format")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrMissingCredentials = errors.New("missing credentials")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
