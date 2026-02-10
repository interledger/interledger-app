# geo

The `geo` package provides types and utilities for handling geographic and monetary data, including countries, currencies, and financial assets.

## Overview

This package contains three core components:

- **Asset** - Represents a monetary asset (e.g., USD, EUR) with ISO 4217 properties
- **Currency** - Handles monetary amounts with arbitrary precision arithmetic
- **Country** - Represents countries with ISO 3166-1 codes and metadata

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

// Check if an asset is supported
if geo.IsSupported("USD") {
    asset, _ := geo.GetAsset("USD")
}

// Get all supported assets
allAssets := geo.AllAssets()
```

**Supported Assets:**
- `USD` - US Dollar (scale: 2)
- `EUR` - Euro (scale: 2)
- `ZAR` - South African Rand (scale: 2)
- `CAD` - Canadian Dollar (scale: 2)
- `JPY` - Japanese Yen (scale: 0)

### Currency

Currency handles monetary amounts with precise arithmetic using `math/big`:

```go
// Create a new currency
amount := geo.NewCurrency(geo.USD())

// Set amount from various formats
amount.SetAmount("123.45")     // string with decimals
amount.SetAmount(123)          // integer (whole units)
amount.SetAmount(int64(123))   // int64

// Get the formatted amount
fmt.Println(amount.Amount()) // "123.45"
fmt.Println(amount.String()) // "$ 123.45"

// Arithmetic operations
other := geo.NewCurrency(geo.USD())
other.SetAmount("50.00")

sum, err := amount.Add(other)     // 173.45
diff, err := amount.Sub(other)    // 73.45
negated := amount.Neg()           // -123.45
absolute := negated.Abs()         // 123.45

// Comparisons
amount.IsZero()                   // false
amount.IsPositive()               // true
amount.IsNegative()               // false
cmp, err := amount.Cmp(other)     // 1 (amount > other)

// Clone for safe modifications
clone := amount.Clone()

// Raw amount access (scaled integer, e.g., cents for USD)
raw := amount.RawAmount()         // *big.Int: 12345 for "123.45" USD
amount.SetRawAmount(big.NewInt(5000)) // Sets to "50.00" USD
```

### Countries

Work with ISO 3166-1 country codes:

```go
// Parse country from string (alpha-2 or numeric)
country := geo.ParseCountry("US")
country = geo.ParseCountry("840")  // Also returns US

// Country properties
fmt.Println(country.String())      // "US"
numeric, _ := country.Numeric()    // "840"
country.Valid()                    // true
country.IsSupported()              // depends on business logic

// EU membership
geo.IsEUCountry(geo.DE)            // true
geo.IsEUCountry(geo.US)            // false
euCountries := geo.EUCountries()   // All 27 EU members

// Country details
detail, ok := geo.GetCountryDetail(geo.US)
fmt.Println(detail.Name)           // "United States of America"

// All countries
allCountries := geo.AllCountries()

// State/Province support
states, ok := geo.GetStates(geo.US)
stateName, ok := geo.GetStateName(geo.US, "CA")  // "CALIFORNIA"

// Alpha-2 to Alpha-3 conversion
alpha3 := geo.ToAlpha3("US")       // "USA"
```

## Protobuf Support

The package includes conversion functions for gRPC/protobuf interoperability:

```go
// Asset <-> Proto
pb := asset.ToProtoGeoV1()
asset, ok := geo.AssetFromProtoGeoV1(pb)

// Currency <-> Proto
pb := currency.ToProtoGeoV1("US")  // includes country code
currency, err := geo.CurrencyFromProtoGeoV1(pb)
```

## Error Handling

The package defines the following errors:

| Error | Description |
|-------|-------------|
| `ErrInvalidFormat` | Invalid amount format when parsing |
| `ErrAssetMismatch` | Arithmetic operation on mismatched assets |
| `ErrUnsupportedType` | Unsupported type passed to SetAmount |
| `ErrUnsupportedAsset` | Asset code not in registry |
| `ErrCountryNotFound` | Country not found |

## Testing

Run tests with coverage:

```bash
go test -v -cover ./...
```

## Thread Safety

- `Asset` is immutable and safe for concurrent use
- `Currency` is **not** thread-safe; use external synchronization or clone for concurrent access
- Country lookup functions are thread-safe (read-only maps)
