package accountdeletion

import "errors"

var (
	ErrInternal         = errors.New("accountdeletion: internal error")
	ErrAlreadyRequested = errors.New("accountdeletion: already requested")
)
