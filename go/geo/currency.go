package geo

import (
	"fmt"
	"math/big"
	"strings"

	geopbv1 "gitlab.com/fynbos/proto/geo/v1"
)

// Currency represents a monetary unit with a specific Asset and amount.
// The amount is stored as a scaled integer (e.g., cents for USD).
type Currency struct {
	amount *big.Int
	asset  Asset
}

// NewCurrency creates a new Currency instance for the given Asset with an initial amount of zero.
func NewCurrency(a Asset) *Currency {
	return &Currency{
		amount: big.NewInt(0),
		asset:  a,
	}
}

// Clone returns a deep copy of the Currency.
func (c *Currency) Clone() *Currency {
	return &Currency{
		amount: new(big.Int).Set(c.amount),
		asset:  c.asset,
	}
}

// Scale returns the number of decimal places for the currency.
func (c *Currency) Scale() uint8 {
	return c.asset.Scale()
}

// Code returns the ISO 4217 code of the currency.
func (c *Currency) Code() string {
	return c.asset.Code()
}

// NumericCode returns the ISO 4217 numeric code of the currency.
func (c *Currency) NumericCode() string {
	return c.asset.NumericCode()
}

// Factor returns the scaling factor for the currency based on its scale.
func (c *Currency) Factor() *big.Int {
	return c.asset.Factor()
}

// RawAmount returns the internal scaled amount as a new big.Int.
// For USD with scale 2, amount "123.45" returns 12345.
func (c *Currency) RawAmount() *big.Int {
	return new(big.Int).Set(c.amount)
}

// SetRawAmount sets the internal scaled amount directly.
// For USD with scale 2, value 12345 represents "123.45".
func (c *Currency) SetRawAmount(value *big.Int) *Currency {
	c.amount = new(big.Int).Set(value)
	return c
}

// IsZero returns true if the amount is zero.
func (c *Currency) IsZero() bool {
	return c.amount.Sign() == 0
}

// IsNegative returns true if the amount is negative.
func (c *Currency) IsNegative() bool {
	return c.amount.Sign() < 0
}

// IsPositive returns true if the amount is positive.
func (c *Currency) IsPositive() bool {
	return c.amount.Sign() > 0
}

// Cmp compares two Currency values. Returns -1 if c < other, 0 if equal, 1 if c > other.
// Returns an error if the assets don't match.
func (c *Currency) Cmp(other *Currency) (int, error) {
	if !c.asset.Equal(other.asset) {
		return 0, ErrAssetMismatch
	}
	return c.amount.Cmp(other.amount), nil
}

// Equal returns true if two Currency values are equal (same asset and amount).
func (c *Currency) Equal(other *Currency) bool {
	return c.asset.Equal(other.asset) && c.amount.Cmp(other.amount) == 0
}

// Add returns a new Currency with the sum of c and other.
// Returns an error if the assets don't match.
func (c *Currency) Add(other *Currency) (*Currency, error) {
	if !c.asset.Equal(other.asset) {
		return nil, ErrAssetMismatch
	}
	result := c.Clone()
	result.amount.Add(c.amount, other.amount)
	return result, nil
}

// Sub returns a new Currency with the difference of c and other.
// Returns an error if the assets don't match.
func (c *Currency) Sub(other *Currency) (*Currency, error) {
	if !c.asset.Equal(other.asset) {
		return nil, ErrAssetMismatch
	}
	result := c.Clone()
	result.amount.Sub(c.amount, other.amount)
	return result, nil
}

// Neg returns a new Currency with the negated amount.
func (c *Currency) Neg() *Currency {
	result := c.Clone()
	result.amount.Neg(c.amount)
	return result
}

// Abs returns a new Currency with the absolute value of the amount.
func (c *Currency) Abs() *Currency {
	result := c.Clone()
	result.amount.Abs(c.amount)
	return result
}

// Amount returns the amount as a human-readable string representation.
// For example, for USD with scale 2 and amount 12345, returns "123.45"
// For JPY with scale 0 and amount 500, returns "500"
func (c *Currency) Amount() string {
	factor := c.Factor()
	abs := new(big.Int).Abs(c.amount)
	whole := new(big.Int).Div(abs, factor)

	sign := ""
	if c.amount.Sign() < 0 {
		sign = "-"
	}

	// For scale 0 currencies (like JPY), return without decimal point
	if c.Scale() == 0 {
		return fmt.Sprintf("%s%s", sign, whole.String())
	}

	fractional := new(big.Int).Mod(abs, factor)
	fractionalStr := fractional.String()
	// Pad fractional part with leading zeros if necessary
	if padding := int(c.Scale()) - len(fractionalStr); padding > 0 {
		fractionalStr = strings.Repeat("0", padding) + fractionalStr
	}
	return fmt.Sprintf("%s%s.%s", sign, whole.String(), fractionalStr)
}

// SetAmount sets the amount of the monetary unit from various types.
// The value is scaled according to the unit's scale.
// The input is human readable, e.g. for USD with scale 2, input 12.34 sets amount to 1234.
// Supported types: int, int64, big.Int, *big.Int, string.
// Returns nil and an error if parsing fails.
func (c *Currency) SetAmount(value any) (*Currency, error) {
	factor := c.Factor()
	amount := new(big.Int)

	switch v := value.(type) {
	case int:
		amount.Mul(big.NewInt(int64(v)), factor)

	case int64:
		amount.Mul(big.NewInt(v), factor)

	case big.Int:
		amount.Mul(new(big.Int).Set(&v), factor)

	case *big.Int:
		amount.Mul(new(big.Int).Set(v), factor)

	case string:
		parsed, err := c.parseString(v, factor)
		if err != nil {
			return nil, err
		}
		amount = parsed

	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, value)
	}

	c.amount = amount
	return c, nil
}

// parseString parses a string amount and returns the scaled big.Int value.
func (c *Currency) parseString(s string, factor *big.Int) (*big.Int, error) {
	// Reject strings with whitespace
	if strings.ContainsAny(s, " \t\n\r") {
		return nil, fmt.Errorf("%w: contains whitespace: %s", ErrInvalidFormat, s)
	}

	s = strings.ReplaceAll(s, ",", "") // remove thousand separators

	// Try integer-like string first
	if !strings.Contains(s, ".") {
		if intLike, ok := new(big.Int).SetString(s, 10); ok {
			return new(big.Int).Mul(intLike, factor), nil
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, s)
	}

	// Must be decimal format with exactly one dot
	if strings.Count(s, ".") != 1 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, s)
	}

	parts := strings.SplitN(s, ".", 2)
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("%w: empty whole part in %s", ErrInvalidFormat, s)
	}
	if len(parts[1]) == 0 {
		return nil, fmt.Errorf("%w: empty fractional part in %s", ErrInvalidFormat, s)
	}

	negative := parts[0][0] == '-'
	wholePart := parts[0]
	fractionalPart := parts[1]

	whole, ok := new(big.Int).SetString(wholePart, 10)
	if !ok {
		return nil, fmt.Errorf("%w: invalid whole part %s", ErrInvalidFormat, wholePart)
	}

	// Normalize fractional part to match scale
	scale := int(c.Scale())
	if len(fractionalPart) > scale {
		fractionalPart = fractionalPart[:scale] // truncate
	} else if len(fractionalPart) < scale {
		fractionalPart += strings.Repeat("0", scale-len(fractionalPart)) // pad
	}

	// Handle negative sign for fractional part
	if negative {
		fractionalPart = "-" + fractionalPart
	}

	fractional, ok := new(big.Int).SetString(fractionalPart, 10)
	if !ok {
		return nil, fmt.Errorf("%w: invalid fractional part %s", ErrInvalidFormat, fractionalPart)
	}

	result := new(big.Int).Mul(whole, factor)
	result.Add(result, fractional)
	return result, nil
}

// String returns the string representation of the currency amount with formatting.
func (c Currency) String() string {
	return c.asset.Format(fmt.Sprintf("%s", c.Amount()))
}

// ToProtoGeoV1 converts the Currency to its protobuf representation.
// The countryCode parameter is optional - pass an empty string if not applicable.
func (c *Currency) ToProtoGeoV1(countryCode string) *geopbv1.Currency {
	return &geopbv1.Currency{
		Amount:      c.amount.String(),
		Asset:       c.asset.ToProtoGeoV1(),
		CountryCode: countryCode,
	}
}

// CurrencyFromProtoGeoV1 creates a Currency from its protobuf representation.
// Returns the Currency and an error if the asset is not supported or amount is invalid.
func CurrencyFromProtoGeoV1(pb *geopbv1.Currency) (*Currency, error) {
	if pb == nil {
		return nil, fmt.Errorf("nil proto currency")
	}
	asset, ok := AssetFromProtoGeoV1(pb.Asset)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAsset, pb.Asset.GetCode())
	}
	currency := NewCurrency(asset)
	amount, ok := new(big.Int).SetString(pb.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, pb.Amount)
	}
	currency.SetRawAmount(amount)
	return currency, nil
}
