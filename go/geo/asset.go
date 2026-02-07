package geo

import (
	"fmt"
	"math/big"

	geopbv1 "gitlab.com/fynbos/proto/geo/v1"
)

/*
An Asset represents a specific monetary currency with its properties
such as ISO 4217 code, numeric code, scale (number of decimal places),
and formatting function for string representation.

Asset is immutable after creation. All fields are private with getter methods.
*/
type Asset struct {
	code       string
	numeric    string
	scale      uint8
	factor     *big.Int
	formatFunc func(value string) string
}

// NewAsset creates a new Asset with the given properties.
// If formatFunc is nil, a default formatter is used that just returns the value.
func NewAsset(code string, numeric string, scale uint8, formatFunc func(string) string) Asset {
	if formatFunc == nil {
		formatFunc = func(value string) string { return value }
	}

	// Pre-compute factor
	tenPow := big.NewInt(10)
	scaleBig := big.NewInt(int64(scale))
	factor := new(big.Int).Exp(tenPow, scaleBig, nil)

	return Asset{
		code:       code,
		numeric:    numeric,
		scale:      scale,
		factor:     factor,
		formatFunc: formatFunc,
	}
}

// Code returns the ISO 4217 alphabetic code (e.g., "USD", "EUR").
func (a *Asset) Code() string {
	return a.code
}

// NumericCode returns the ISO 4217 numeric code (e.g., "840" for USD).
func (a *Asset) NumericCode() string {
	return a.numeric
}

// Scale returns the number of decimal places for this asset.
func (a *Asset) Scale() uint8 {
	return a.scale
}

// Factor returns the scaling factor (10^scale) for this asset.
func (a *Asset) Factor() *big.Int {
	return a.factor
}

// Format applies the asset's formatting function to the given value string.
func (a *Asset) Format(value string) string {
	return a.formatFunc(value)
}

// Equal returns true if two assets have the same code.
func (a Asset) Equal(other Asset) bool {
	return a.code == other.code
}

// String returns a string representation of the asset for debugging.
func (a Asset) String() string {
	return fmt.Sprintf("Asset{code=%s, numeric=%s, scale=%d}", a.code, a.numeric, a.scale)
}

// ToProtoGeoV1 converts the Asset to its protobuf representation.
func (a Asset) ToProtoGeoV1() *geopbv1.Asset {
	return &geopbv1.Asset{
		Code:    a.code,
		Numeric: a.numeric,
		Scale:   uint32(a.scale),
	}
}

// AssetFromProtoGeoV1 creates an Asset from its protobuf representation.
// Returns the Asset and a boolean indicating if the asset code is supported.
func AssetFromProtoGeoV1(pb *geopbv1.Asset) (Asset, bool) {
	if pb == nil {
		return Asset{}, false
	}
	return GetAsset(pb.Code)
}

// assets is the internal registry of known currency assets.
var assets = map[string]Asset{
	"USD": NewAsset("USD", "840", 2, func(value string) string { return "$ " + value }),
	"EUR": NewAsset("EUR", "978", 2, func(value string) string { return "€ " + value }),
	"ZAR": NewAsset("ZAR", "710", 2, func(value string) string { return "R " + value }),
	"CAD": NewAsset("CAD", "124", 2, func(value string) string { return "$ " + value }),
	"JPY": NewAsset("JPY", "392", 0, func(value string) string { return "¥ " + value }),
}

// GetAsset returns the Asset for the given code and a boolean indicating if it exists.
// Code matching is case-sensitive. Returns a copy, so the original cannot be modified.
func GetAsset(code string) (Asset, bool) {
	asset, ok := assets[code]
	return asset, ok
}

// IsSupported returns true if the given asset code is supported.
func IsSupported(code string) bool {
	_, ok := assets[code]
	return ok
}

// AllAssets returns a slice of all supported assets.
func AllAssets() []Asset {
	result := make([]Asset, 0, len(assets))
	for _, a := range assets {
		result = append(result, a)
	}
	return result
}

// Convenience accessors for known assets.
func USD() Asset { return assets["USD"] }
func EUR() Asset { return assets["EUR"] }
func ZAR() Asset { return assets["ZAR"] }
func CAD() Asset { return assets["CAD"] }
func JPY() Asset { return assets["JPY"] }
