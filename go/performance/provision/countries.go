package provision

import (
	"fmt"
	"strings"

	"github.com/interledger/interledger-app/go/backend/currency"
)

type countrySpec struct {
	code         string
	country      string
	currencyCode string
	asset        string
	scale        int32
	provider     string
}

var countrySpecs = map[string]countrySpec{
	"za": {code: "za", country: "ZA", currencyCode: "ZAR", asset: "ZAR", scale: 2, provider: "xago"},
	"de": {code: "de", country: "DE", currencyCode: "EUR", asset: "EUR", scale: 2, provider: "gatehub"},
	"us": {code: "us", country: "US", currencyCode: "USD", asset: "USD", scale: 2, provider: "pti"},
}

func parseCountries(raw []string) ([]countrySpec, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("at least one country is required")
	}

	seen := make([]countrySpec, 0, len(raw))
	for _, entry := range raw {
		key := strings.ToLower(strings.TrimSpace(entry))
		if key == "" {
			continue
		}
		spec, ok := countrySpecs[key]
		if !ok {
			return nil, fmt.Errorf("unsupported country %q", entry)
		}
		seen = append(seen, spec)
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("at least one country is required")
	}
	return seen, nil
}

func walletLabel(prefix, countryCode string, i int) string {
	return fmt.Sprintf("%s-%s-%03d", prefix, strings.ToLower(countryCode), i)
}

func phoneNumber(prefix string, globalIndex int) string {
	return fmt.Sprintf("%s%07d", prefix, globalIndex)
}

func walletAddress(host, label string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(host, "/"), label)
}

func minorAmount(targetMajor int64, scale int32) int64 {
	return targetMinorAmount(targetMajor, scale)
}

func targetMinorAmount(targetMajor int64, scale int32) int64 {
	if scale <= 0 {
		return targetMajor
	}

	amount := targetMajor
	for i := int32(0); i < scale; i++ {
		amount *= 10
	}
	return amount
}

func assetScale(currencyCode string) int32 {
	return int32(currency.ParseCurrency(currencyCode).Scale())
}
