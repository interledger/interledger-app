package kyc

import "errors"

var (
	ErrNoKYCInfo               = errors.New("kyc: no data found")
	ErrInternal                = errors.New("kyc: internal error")
	ErrKYCCompleted            = errors.New("kyc: already completed")
	ErrKYCResubmissionRequired = errors.New("kyc resubmission required: please update your verification documents")
	ErrKYCNotApproved          = errors.New("kyc not approved")
)
