package signup

import "errors"

var (
	ErrInternal        = errors.New("internal error")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
	ErrInvalidOTP      = errors.New("invalid OTP")
	ErrDuplicatePhone  = errors.New("duplicate phone number used")
)
