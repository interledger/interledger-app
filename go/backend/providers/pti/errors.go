package pti

import "errors"

var (
	ErrInternal            = errors.New("pti: internal error.")
	ErrAssessmentFailed    = errors.New("pti: assessment failed")
	ErrNotFound            = errors.New("pti: not found.")
	ErrInsufficientBalance = errors.New("pti: insufficient balance")
	ErrTimedOut            = errors.New("pti: timed out")
)
