package currency

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"

	"gitlab.com/fynbos/backend/country"
	pb "gitlab.com/fynbos/proto/backend/v1"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Currency string

func ParseCurrency(cc string) Currency {
	formattedInput := strings.ToUpper(strings.TrimSpace(cc))
	_, ok := iso4217[Currency(formattedInput)]
	if ok {
		return Currency(formattedInput)
	}

	ret, ok := iso4217Currency[formattedInput]
	if ok {
		return ret
	}

	return Currency(formattedInput)
}

func FromISO4217(iso string) Currency {
	cc, ok := iso4217Currency[iso]
	if !ok {
		log.Warn("unknown iso4217 code for currency", zap.String("iso4217", iso))
		return ""
	}
	return cc
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
	EUR Currency = "EUR"
	ZAR Currency = "ZAR"
	JPY Currency = "JPY"
	GBP Currency = "GBP"
	INR Currency = "INR"
	CAD Currency = "CAD"
)

var currencyScale = map[Currency]int{
	USD: 2,
	EUR: 2,
	ZAR: 2,
	JPY: 0,
	GBP: 2,
	INR: 2,
	CAD: 2,
}

var currencyFormat = map[Currency]string{
	USD: "$ %s",
	EUR: "€ %s",
	ZAR: "R %s",
	JPY: "%s ¥",
	GBP: "£ %s",
	INR: "₹ %s",
	CAD: "$ %s",
}

var iso4217 = map[Currency]string{
	USD: "840",
	EUR: "978",
	ZAR: "710",
	JPY: "392",
	GBP: "826",
	INR: "356",
	CAD: "124",
}

var iso4217Currency = map[string]Currency{
	"840": USD,
	"978": EUR,
	"710": ZAR,
	"392": JPY,
	"826": GBP,
	"356": INR,
	"124": CAD,
}

type Amount struct {
	Value    uint64   `validate:"gte=0" json:"amount,string"`
	Currency Currency `validate:"iso4217"  json:"currency"`
	Scale    int      `validate:"omitempty,gte=0" json:"scale"`
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

	cntry := country.ParseCountry(a.Currency.ISO4217()).String()
	// EUR is a special case where we want to show the Euro flag on the frontend
	if a.Currency == EUR {
		cntry = "EU"
	}

	return &pb.Amount{
		Amount:     a.Value,
		Asset:      a.Currency.String(),
		AssetScale: int32(a.Scale),
		Country:    cntry,
	}
}

func (a *Amount) ToAdminPB() *adminv1.Amount {
	if a == nil {
		return nil
	}

	cntry := country.ParseCountry(a.Currency.ISO4217()).String()
	// EUR is a special case where we want to show the Euro flag on the frontend
	if a.Currency == EUR {
		cntry = "EU"
	}

	return &adminv1.Amount{
		Amount:     a.Value,
		Asset:      a.Currency.String(),
		AssetScale: int32(a.Scale),
		Country:    cntry,
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
	smallest := math.Pow(10, -float64(cc.Scale()))

	convertedAmount := f * math.Pow(10, float64(cc.Scale()))
	// start with truncated value
	amt := uint64(convertedAmount)

	// use the rounded value if it's close enough
	if math.Abs(math.Round(convertedAmount)-convertedAmount) < smallest {
		amt = uint64(math.Round(convertedAmount))
	}

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

func FromString(value string, cc Currency) (Amount, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return Amount{}, err
	}

	return FromFloat64(f, cc), nil
}

func StringToScaledUInt(input string) (uint64, error) {
	input = strings.ReplaceAll(input, ",", "")
	if len(input) == 0 {
		return 0, errors.New("empty string provided")
	}

	parts := strings.Split(input, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return 0, errors.New("invalid format should be numeric or numeric.numeric")
	}

	beforeDotValue, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid string before decimal point should be numeric")
	}

	var afterDotValue uint64 = 0
	if len(parts) == 2 {
		afterDotValue, err = strconv.ParseUint(parts[1], 10, 64)
		if len(parts[1]) == 1 {
			afterDotValue = afterDotValue * 10
		}
		if err != nil {
			return 0, errors.New("invalid string after decimal point should be numeric")
		}
	}
	fee := uint64(beforeDotValue*100 + afterDotValue)

	return fee, nil
}
