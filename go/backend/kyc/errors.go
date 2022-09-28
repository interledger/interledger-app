package kyc

import "errors"

var (
	ErrNoKYCInfo = errors.New("no KYC data found")
	ErrInternal  = errors.New("internal KYC error")
)
