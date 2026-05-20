package user

import "errors"

var (
	ErrInternal          = errors.New("user: internal error")
	ErrInvalidArgument   = errors.New("user: invalid argument")
	ErrNoUserFound       = errors.New("no user found")
	ErrNoCredentials     = errors.New("no credentials found")
	ErrTotpNotConfigured = errors.New("totp not configured")
	ErrInvalidTotpConfig = errors.New("invalid totp config")
	ErrAAL1Required      = errors.New("aal1 authentication required")
	ErrAAL2Required      = errors.New("aal2 authentication required")
)
