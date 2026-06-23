package email

import "errors"

var (
	ErrInvalidTemplate           = errors.New("email: invalid template")
	ErrInternal                  = errors.New("email: internal error")
	ErrSupportInboxNotConfigured = errors.New("email: support inbox not configured")
)
