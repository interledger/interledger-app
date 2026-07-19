package geo

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	geopbv1 "github.com/interledger/interledger-app/go/proto/geo/v1"
)

// Currency represents a monetary unit with a specific Asset and amount.
// The amount is stored as a scaled integer (e.g., cents for USD).
type Currency struct {
	amount big.Int
	asset  Asset
}

// NewCurrency creates a new Currency instance for the given Asset with an initial amount of zero.
func NewCurrency(a Asset) *Currency {
	return &Currency{
		asset: a,
	}
}

// Clone returns a deep copy of the Currency.
func (c *Currency) Clone() *Currency {
	return &Currency{
		amount: *new(big.Int).Set(&c.amount),
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
	return new(big.Int).Set(&c.amount)
}

// SetRawAmountBigInt sets the internal scaled amount directly from a *big.Int.
// For USD with scale 2, value 12345 represents "123.45".
func (c *Currency) SetRawAmountBigInt(value *big.Int) {
	c.amount.Set(value)
}

// SetRawAmountInt64 sets the internal scaled amount directly from an int64.
// For USD with scale 2, value 12345 represents "123.45".
func (c *Currency) SetRawAmountInt64(value int64) {
	c.amount.SetInt64(value)
}

// SetRawAmountUint64 sets the internal scaled amount directly from a uint64.
// For USD with scale 2, value 12345 represents "123.45".
func (c *Currency) SetRawAmountUint64(value uint64) {
	c.amount.SetUint64(value)
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
	return c.amount.Cmp(&other.amount), nil
}

// Equal returns true if two Currency values are equal (same asset and amount).
func (c *Currency) Equal(other *Currency) bool {
	return c.asset.Equal(other.asset) && c.amount.Cmp(&other.amount) == 0
}

// Add adds other's amount to c in place.
// Returns an error if the assets don't match.
func (c *Currency) Add(other *Currency) error {
	if !c.asset.Equal(other.asset) {
		return ErrAssetMismatch
	}
	c.amount.Add(&c.amount, &other.amount)
	return nil
}

// Sub subtracts other's amount from c in place.
// Returns an error if the assets don't match.
func (c *Currency) Sub(other *Currency) error {
	if !c.asset.Equal(other.asset) {
		return ErrAssetMismatch
	}
	c.amount.Sub(&c.amount, &other.amount)
	return nil
}

// Neg negates the amount in place.
func (c *Currency) Neg() {
	c.amount.Neg(&c.amount)
}

// Abs sets the amount to its absolute value in place.
func (c *Currency) Abs() {
	c.amount.Abs(&c.amount)
}

// Amount returns the amount as a human-readable string representation.
// For example, for USD with scale 2 and amount 12345, returns "123.45"
// For JPY with scale 0 and amount 500, returns "500"
func (c *Currency) Amount() string {
	factor := &c.asset.factor
	abs := new(big.Int).Abs(&c.amount)

	sign := ""
	if c.amount.Sign() < 0 {
		sign = "-"
	}

	// For scale 0 currencies (like JPY), return without decimal point
	if c.Scale() == 0 {
		return sign + abs.String()
	}

	fractional := new(big.Int)
	abs.DivMod(abs, factor, fractional)
	// abs now holds the whole part

	fractionalStr := fractional.String()
	// Pad fractional part with leading zeros if necessary
	if padding := int(c.Scale()) - len(fractionalStr); padding > 0 {
		fractionalStr = strings.Repeat("0", padding) + fractionalStr
	}
	return sign + abs.String() + "." + fractionalStr
}

// SetAmountString sets the amount from a human-readable string representation.
// The value is scaled according to the asset's scale.
// For USD with scale 2, input "12.34" sets the internal amount to 1234.
// So far, this is the only way to set the human-readable amount.
func (c *Currency) SetAmountString(s string) error {
	parsed, err := c.parseString(s, &c.asset.factor)
	if err != nil {
		return err
	}
	c.amount = *parsed
	return nil
}

// parseString parses a string amount and returns the scaled big.Int value.
func (c *Currency) parseString(s string, factor *big.Int) (*big.Int, error) {
	// Reject strings with whitespace
	if strings.ContainsAny(s, " \t\n\r") {
		return nil, fmt.Errorf("%w: contains whitespace: %s", ErrInvalidFormat, s)
	}

	if strings.ContainsRune(s, ',') {
		if err := validateCommas(s); err != nil {
			return nil, err
		}
		s = strings.ReplaceAll(s, ",", "")
	}

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

// validateCommas checks that commas are correctly placed as thousand separators.
// Valid: "1,234", "12,345", "1,234,567", "-1,234.56"
// Invalid: ",123", "1,,234", "1,23", "1234,567", "1.234,56"
func validateCommas(s string) error {
	// Commas must only appear in the whole part
	whole := s
	if dotIdx := strings.IndexByte(s, '.'); dotIdx != -1 {
		whole = s[:dotIdx]
		if strings.ContainsRune(s[dotIdx:], ',') {
			return fmt.Errorf("%w: commas in fractional part: %s", ErrInvalidFormat, s)
		}
	}

	groups := strings.Split(whole, ",")

	// Strip optional sign from first group
	first := groups[0]
	if len(first) > 0 && (first[0] == '+' || first[0] == '-') {
		first = first[1:]
	}
	if len(first) == 0 || len(first) > 3 {
		return fmt.Errorf("%w: invalid comma placement: %s", ErrInvalidFormat, s)
	}
	for _, g := range groups[1:] {
		if len(g) != 3 {
			return fmt.Errorf("%w: invalid comma placement: %s", ErrInvalidFormat, s)
		}
	}
	return nil
}

// String returns the string representation of the currency amount with formatting.
func (c *Currency) String() string {
	return c.asset.Format(c.Amount())
}

type currencyJSON struct {
	Asset  Asset  `json:"asset"`
	Amount string `json:"amount"`
}

// MarshalJSON implements json.Marshaler for Currency.
// Serializes as {"asset":"USD","amount":"123.45"}.
func (c *Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(currencyJSON{
		Asset:  c.asset,
		Amount: c.Amount(),
	})
}

// UnmarshalJSON implements json.Unmarshaler for Currency.
// Expects {"asset":"USD","amount":"123.45"}.
func (c *Currency) UnmarshalJSON(data []byte) error {
	var j currencyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.asset = j.Asset
	return c.SetAmountString(j.Amount)
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
	currency.SetRawAmountBigInt(amount)
	return currency, nil
}
