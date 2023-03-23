package verygoodsecurity

import "errors"

var (
	ErrInternal            = errors.New("verygoodsecurity: internal error")
	ErrInvalidArgument     = errors.New("verygoodsecurity: invalid argument")
	ErrUserHasExistingCard = errors.New("verygoodsecurity: user has an existing card")
)
