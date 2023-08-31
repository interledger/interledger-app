package discord

import "errors"

var (
	ErrInternal = errors.New("discord: internal error")
	ErrNotFound = errors.New("discord: not found")
)
