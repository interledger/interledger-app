package chimoney

import "errors"

var (
	ErrInternal             = errors.New("chimoney: internal error")
	ErrNotFound             = errors.New("chimoney: not found")
	ErrInsufficientBalance  = errors.New("chimoney: insufficient balance")
	ErrInteracAlreadyLinked = errors.New("chimoney: interac account already linked")
)
