package kyc

import "errors"

var (
	ErrNoKYCInfo = errors.New("kyc: no data found")
	ErrInternal  = errors.New("kyc: internal error")
)
