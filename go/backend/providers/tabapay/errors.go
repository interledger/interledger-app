package tabapay

import "errors"

var (
	ErrInternal    = errors.New("tabapay: internal error.")
	ErrInvalidCard = errors.New("tabapay: invalid card.")
)
