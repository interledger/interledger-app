package linkedin

import "errors"

var (
	ErrInternal         = errors.New("linkedin: internal error")
	ErrNotFound         = errors.New("linkedin: not found")
	ErrConnectionExists = errors.New("linkedin: connection already exists")
)
