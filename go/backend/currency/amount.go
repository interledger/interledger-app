package currency

import (
	"fmt"
	"math"
	"strconv"

	pb "gitlab.com/fynbos/proto/backend/v1"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Currency string

func ParseCurrency(cc string) Currency {
	return Currency(cc)
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) Valid() bool {
	_, ok := currencyScale[c]
	return ok
}

func (c Currency) Scale() int {
	scale, ok := currencyScale[c]
	if !ok {
		log.Warn("unknown scale for currency", zap.String("currency", string(c)))
		return 2 // Default to cent scale
	}

	return scale
}

func (c Currency) ISO4217() string {
	code, ok := iso4217[c]
	if !ok {
		log.Warn("unknown iso4217 code for currency", zap.String("currency", string(c)))
		return ""
	}

	return code
}

const (
	USD Currency = "USD"
)

var currencyScale = map[Currency]int{
	USD: 2,
}

var currencyFormat = map[Currency]string{
	USD: "$ %s",
}

var iso4217 = map[Currency]string{
	USD: "840",
}

type Amount struct {
	Value    uint64   `validate:"gt=0" json:"amount,string"`
	Currency Currency `validate:"iso4217"  json:"currency"`
	Scale    int      `validate:"omitempty,gt=0" json:"scale"`
}

// IsEqual returns true if the value and currency are the same
func (a *Amount) IsEqual(amt Amount) bool {
	if a == nil {
		return false
	}
	return a.Value == amt.Value && a.Currency == amt.Currency
}

// IsEmpty returns true if the value is 0 and currency is an empty string
func (a *Amount) IsEmpty() bool {
	if a == nil {
		return true
	}
	return a.Value == 0 && a.Currency.String() == ""
}

// FormatAmount returns the amount in the scale as representative of a human readable format.
// i.e. Amount { value: 1000, scale: 2} returns "10.00"
func (a *Amount) FormatAmount() string {
	if a == nil {
		return ""
	}
	scale := a.Scale
	if a.Scale <= 0 {
		// Fallback to double-check that the scale actually is 0
		scale = a.Currency.Scale()
	}

	return strconv.FormatFloat(a.Float64(), 'f', scale, 64)
}

// Format returns the amount with its specified formatting
// i.e. Amount { value: 1000, scale: 2, currency: USD} returns "$10.00"
func (a *Amount) Format() string {
	if a == nil {
		return ""
	}

	format, ok := currencyFormat[a.Currency]
	if !ok {
		format = "%s " + a.Currency.String()
	}

	return fmt.Sprintf(format, a.FormatAmount())
}

// Float64 returns the floating point representation of the amount in the scale defined
// i.e. Amount { value: 1000, scale: 2, currency: USD} returns 10.0
func (a *Amount) Float64() float64 {
	if a == nil {
		return 0.00
	}

	scale := float64(a.Scale)
	if a.Scale <= 0 {
		// Fallback to double-check that the scale actually is 0
		scale = float64(a.Currency.Scale())
	}

	amnt := float64(a.Value)
	if scale > 0 {
		amnt /= math.Pow(10, scale)
	}

	return amnt
}

func (a *Amount) ToPB() *pb.Amount {
	if a == nil {
		return nil
	}
	return &pb.Amount{
		Amount:     a.Value,
		Asset:      a.Currency.String(),
		AssetScale: int32(a.Scale),
	}
}

func FromPB(a *pb.Amount) Amount {
	return Amount{
		Value:    a.GetAmount(),
		Currency: Currency(a.GetAsset()),
		Scale:    int(a.GetAssetScale()),
	}
}

func FromFloat64(f float64, cc Currency) Amount {
	amt := uint64(f * math.Pow(10, float64(cc.Scale())))

	return Amount{
		Value:    amt,
		Currency: cc,
		Scale:    cc.Scale(),
	}
}

func FromUInt64(i uint64, cc Currency) Amount {
	return Amount{
		Value:    i,
		Currency: cc,
		Scale:    cc.Scale(),
	}
}
