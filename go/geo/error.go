package geo

import "errors"

var (
	ErrInvalidFormat    = errors.New("invalid amount format")
	ErrAssetMismatch    = errors.New("currency assets do not match")
	ErrUnsupportedAsset = errors.New("unsupported asset")

	ErrCountryNotFound = errors.New("country not found")
)
