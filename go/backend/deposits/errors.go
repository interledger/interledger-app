package deposits

import "errors"

var (
	ErrUnauthorized            = errors.New("deposit: unauthorized")
	ErrInternal                = errors.New("deposit: internal error")
	ErrInvalidArgument         = errors.New("deposit: invalid argument")
	ErrNotFound                = errors.New("deposit: not found")
	ErrUnverifiedFundingSource = errors.New("deposit: unverified funding source")
)
