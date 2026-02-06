package geo

import "errors"

var (
	ErrInvalidFormat   = errors.New("invalid amount format")
	ErrAssetMismatch   = errors.New("currency assets do not match")
	ErrUnsupportedType = errors.New("unsupported type for SetAmount")

	ErrNotFound = errors.New("country: not found.")
)
