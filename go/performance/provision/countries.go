package provision

import (
	"fmt"
	"strconv"
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

func phoneRangeForCountry(country string) (prefix string, digits int) {
	switch strings.ToLower(country) {
	case "za", "south africa":
		return "+27710", 6
	case "us", "usa", "united states":
		return "+1202555", 4
	default:
		return "+491700", 6
	}
}

func phoneNumber(country string, globalIndex int, overridePrefix string, seed int64) string {
	if overridePrefix != "" {
		return fmt.Sprintf("%s%07d", overridePrefix, globalIndex)
	}

	prefix, digits := phoneRangeForCountry(country)
	max := int64(1)
	for i := 0; i < digits; i++ {
		max *= 10
	}

	suffix := strconv.FormatInt((int64(globalIndex)+seed)%max, 10)
	return prefix + strings.Repeat("0", digits-len(suffix)) + suffix
}

func walletAddressLabel(label string) string {
	clean := strings.ToLower(strings.TrimSpace(label))
	var b strings.Builder
	for _, r := range clean {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_':
			b.WriteRune(r)
		default:
			// Drop separators such as - or . so the resulting label stays valid.
		}
	}
	clean = strings.Trim(b.String(), "_")
	if clean == "" {
		clean = "perf"
	}
	if len(clean) < 3 {
		clean = "perf" + clean
	}
	if len(clean) > 16 {
		clean = clean[:16]
	}
	if clean != "" && clean[len(clean)-1] == '_' {
		clean = clean[:len(clean)-1]
	}
	return clean
}

func walletAddress(host, label string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(host, "/"), walletAddressLabel(label))
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
