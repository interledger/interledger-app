# geo

The `geo` package provides types and utilities for handling geographic and monetary data, including countries, currencies, and financial assets.

## Overview

This package contains three core components:

- **Asset** — Represents a monetary asset (e.g., USD, EUR) with ISO 4217 properties
- **Currency** — Handles monetary amounts with arbitrary-precision arithmetic via `math/big`
- **Country** — Represents countries with ISO 3166-1 codes and metadata

## Installation

```go
import "gitlab.com/fynbos/geo"
```

## Usage

### Assets

Assets represent monetary currencies with their ISO 4217 properties:

```go
// Use predefined assets
usd := geo.USD()
eur := geo.EUR()
jpy := geo.JPY()

// Access asset properties
fmt.Println(usd.Code())        // "USD"
fmt.Println(usd.NumericCode()) // "840"
fmt.Println(usd.Scale())       // 2 (decimal places)
fmt.Println(usd.Factor())      // 100 (10^scale)
fmt.Println(usd.Format("9.99"))// "$ 9.99"

// Create a custom asset
custom := geo.NewAsset("XTS", "963", 4, func(v string) string { return v + " XTS" })

// Check if an asset is in the registry
if geo.IsSupported("USD") {
    asset, _ := geo.GetAsset("USD")
}

// Get all registered assets
allAssets := geo.AllAssets()

// Equality check (compares by code)
usd.Equal(eur) // false
```

**Registered Assets:**

| Code | Name | Numeric | Scale | Symbol |
|------|------|---------|-------|--------|
| `USD` | US Dollar | 840 | 2 | `$` |
| `EUR` | Euro | 978 | 2 | `€` |
| `ZAR` | South African Rand | 710 | 2 | `R` |
| `CAD` | Canadian Dollar | 124 | 2 | `$` |
| `JPY` | Japanese Yen | 392 | 0 | `¥` |

### Currency

Currency handles monetary amounts with precise arithmetic using `math/big`. All arithmetic operations (`Add`, `Sub`, `Neg`, `Abs`) **mutate the receiver in place**.

```go
// Create a new currency (initial amount is zero)
amount := geo.NewCurrency(geo.USD())

// Set amount from a string (supports decimals and comma-separated thousands)
amount.SetAmount("1,234.56")

// Set amount from numeric types (value represents whole units, scaled internally)
amount.SetAmountInt(123)                  // internal: 12300 (cents)
amount.SetAmountUint64(123)               // internal: 12300
amount.SetAmountBigInt(big.NewInt(123))   // internal: 12300

// Read the amount
fmt.Println(amount.Amount()) // "123.00"
fmt.Println(amount.String()) // "$ 123.00"
fmt.Println(amount.Code())   // "USD"
fmt.Println(amount.Scale())  // 2

// Arithmetic (in-place — use Clone() first if you need the original)
other := geo.NewCurrency(geo.USD())
other.SetAmount("50.00")

err := amount.Add(other)   // amount is now 173.00
err = amount.Sub(other)    // amount is now 123.00
amount.Neg()               // amount is now -123.00
amount.Abs()               // amount is now 123.00

// Comparisons
amount.IsZero()                 // false
amount.IsPositive()             // true
amount.IsNegative()             // false
amount.Equal(other)             // false
cmp, err := amount.Cmp(other)   // 1 (amount > other)

// Clone for safe modifications
clone := amount.Clone()

// Raw amount access (scaled integer, e.g., cents for USD)
raw := amount.RawAmount()              // *big.Int: 12300 for "123.00" USD
amount.SetRawAmount(big.NewInt(5000))  // Sets to "50.00" USD
```

> **Note:** `Add`, `Sub`, `Neg`, and `Abs` modify the receiver. `Add` and `Sub` return an `error` (non-nil when assets don't match); `Neg` and `Abs` have no return value.

### Countries

Work with ISO 3166-1 country codes:

```go
// Parse country from string (alpha-2 or numeric; defaults to US on invalid input)
country := geo.ParseCountry("US")
country = geo.ParseCountry("840")  // Also returns US

// Country properties
fmt.Println(country.String())      // "US"
numeric, _ := country.Numeric()    // "840"
country.Valid()                     // true
country.IsSupported()              // true if marked as supported in the registry

// EU membership
geo.IsEUCountry(geo.DE)            // true
geo.IsEUCountry(geo.US)            // false
euCountries := geo.EUCountries()   // All 27 EU member states

// Country details
detail, ok := geo.GetCountryDetail(geo.US)
fmt.Println(detail.Name)           // "United States of America"
fmt.Println(detail.Numeric)        // "840"
fmt.Println(detail.Supported)      // true

// All countries (249 entries)
allCountries := geo.AllCountries()

// State/Province support (US and GB)
states, ok := geo.GetStates(geo.US)         // map of state code → name
stateName, ok := geo.GetStateName(geo.US, "CA")  // "CALIFORNIA"

// Alpha-2 to Alpha-3 conversion
alpha3 := geo.ToAlpha3("US")       // "USA"
```

## Protobuf Support

All core types include conversion functions for gRPC/protobuf interoperability via `proto/geo/v1`:

```go
// Asset <-> Proto
pb := asset.ToProtoGeoV1()
asset, ok := geo.AssetFromProtoGeoV1(pb)

// Currency <-> Proto (includes an optional country code)
pb := currency.ToProtoGeoV1("US")
currency, err := geo.CurrencyFromProtoGeoV1(pb)
```

## Error Handling

The package defines the following sentinel errors:

| Error | Description |
|-------|-------------|
| `ErrInvalidFormat` | Invalid amount format when parsing a string |
| `ErrAssetMismatch` | Arithmetic operation on currencies with different assets |
| `ErrUnsupportedAsset` | Asset code not found in the registry |
| `ErrCountryNotFound` | Country code not found in the details map |

## Testing

Run tests with coverage:

```bash
go test -v -cover ./geo/...
```

## Thread Safety

- `Asset` is immutable after creation and safe for concurrent use
- `Currency` is **not** thread-safe; use `Clone()` or external synchronization for concurrent access
- Country lookup functions are thread-safe (read-only package-level maps)
