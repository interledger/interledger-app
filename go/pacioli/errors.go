package pacioli

import "errors"

var (
	ErrInvalidArg = errors.New("pacioli: invalid argument.")
	ErrNotFound   = errors.New("pacioli: not found.")
	ErrDuplicate  = errors.New("pacioli: duplicate.")
	ErrInternal   = errors.New("pacioli: internal error.")
)
