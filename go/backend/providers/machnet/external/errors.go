package external

import "errors"

var (
	ErrNotFound        = errors.New("machnet external: not found.")
	ErrInternal        = errors.New("machnet external: internal error.")
	ErrInvalidArgument = errors.New("machnet external: invalid argument.")
)
