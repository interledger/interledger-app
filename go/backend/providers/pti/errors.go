package pti

import "errors"

var (
	ErrInternal         = errors.New("pti: internal error.")
	ErrAssessmentFailed = errors.New("pti: assessment failed")
)
