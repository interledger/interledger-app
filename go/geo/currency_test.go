package geo

import (
	"math/big"
	"testing"

	geopbv1 "gitlab.com/fynbos/proto/geo/v1"
)

func TestClone(t *testing.T) {
	original := NewCurrency(USD())
	original.SetAmountString("100.00")

	clone := original.Clone()

	// Verify clone has same value
	if clone.Amount() != original.Amount() {
		t.Errorf("Clone amount %s != original amount %s", clone.Amount(), original.Amount())
	}

	// Modify clone and verify original is unchanged
	clone.SetAmountString("200.00")
	if original.Amount() != "100.00" {
		t.Errorf("Original was modified when clone changed: got %s, want 100.00", original.Amount())
	}
}

func TestRawAmount(t *testing.T) {
	currency := NewCurrency(USD())
	currency.SetAmountString("123.45")

	raw := currency.RawAmount()
	expected := big.NewInt(12345)

	if raw.Cmp(expected) != 0 {
		t.Errorf("RawAmount: got %s, want %s", raw.String(), expected.String())
	}

	// Verify returned value is a copy (modifying it doesn't affect original)
	raw.SetInt64(99999)
	if currency.RawAmount().Cmp(big.NewInt(12345)) != 0 {
		t.Error("RawAmount returned reference instead of copy")
	}
}

func TestSetRawAmountBigInt(t *testing.T) {
	currency := NewCurrency(USD())
	currency.SetRawAmountBigInt(big.NewInt(12345))

	if currency.Amount() != "123.45" {
		t.Errorf("SetRawAmountBigInt: got %s, want 123.45", currency.Amount())
	}

	// Verify it creates a copy (modifying input doesn't affect currency)
	input := big.NewInt(50000)
	currency.SetRawAmountBigInt(input)
	input.SetInt64(0)
	if currency.RawAmount().Cmp(big.NewInt(50000)) != 0 {
		t.Error("SetRawAmountBigInt used reference instead of copy")
	}
}

func TestSetRawAmountInt64(t *testing.T) {
	currency := NewCurrency(USD())
	currency.SetRawAmountInt64(12345)

	if currency.Amount() != "123.45" {
		t.Errorf("SetRawAmountInt64: got %s, want 123.45", currency.Amount())
	}

	currency.SetRawAmountInt64(-6789)
	if currency.Amount() != "-67.89" {
		t.Errorf("SetRawAmountInt64 negative: got %s, want -67.89", currency.Amount())
	}

	currency.SetRawAmountInt64(0)
	if currency.Amount() != "0.00" {
		t.Errorf("SetRawAmountInt64 zero: got %s, want 0.00", currency.Amount())
	}
}

func TestSetRawAmountUint64(t *testing.T) {
	currency := NewCurrency(USD())
	currency.SetRawAmountUint64(12345)

	if currency.Amount() != "123.45" {
		t.Errorf("SetRawAmountUint64: got %s, want 123.45", currency.Amount())
	}

	currency.SetRawAmountUint64(0)
	if currency.Amount() != "0.00" {
		t.Errorf("SetRawAmountUint64 zero: got %s, want 0.00", currency.Amount())
	}
}

func TestIsZero(t *testing.T) {
	cases := []struct {
		amount   string
		expected bool
	}{
		{"0", true},
		{"0.00", true},
		{"-0", true},
		{"0.01", false},
		{"-0.01", false},
		{"100", false},
		{"-100", false},
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		currency.SetAmountString(test.amount)
		if currency.IsZero() != test.expected {
			t.Errorf("IsZero(%s): got %v, want %v", test.amount, currency.IsZero(), test.expected)
		}
	}
}

func TestIsNegative(t *testing.T) {
	cases := []struct {
		amount   string
		expected bool
	}{
		{"0", false},
		{"100", false},
		{"0.01", false},
		{"-0.01", true},
		{"-100", true},
		{"-0", false}, // negative zero is still zero
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		currency.SetAmountString(test.amount)
		if currency.IsNegative() != test.expected {
			t.Errorf("IsNegative(%s): got %v, want %v", test.amount, currency.IsNegative(), test.expected)
		}
	}
}

func TestIsPositive(t *testing.T) {
	cases := []struct {
		amount   string
		expected bool
	}{
		{"0", false},
		{"100", true},
		{"0.01", true},
		{"-0.01", false},
		{"-100", false},
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		currency.SetAmountString(test.amount)
		if currency.IsPositive() != test.expected {
			t.Errorf("IsPositive(%s): got %v, want %v", test.amount, currency.IsPositive(), test.expected)
		}
	}
}

func TestCmp(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("100.00")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("200.00")

	usd3 := NewCurrency(USD())
	usd3.SetAmountString("100.00")

	// Less than
	cmp, err := usd1.Cmp(usd2)
	if err != nil {
		t.Fatalf("Cmp failed: %v", err)
	}
	if cmp != -1 {
		t.Errorf("100 < 200: got %d, want -1", cmp)
	}

	// Greater than
	cmp, err = usd2.Cmp(usd1)
	if err != nil {
		t.Fatalf("Cmp failed: %v", err)
	}
	if cmp != 1 {
		t.Errorf("200 > 100: got %d, want 1", cmp)
	}

	// Equal
	cmp, err = usd1.Cmp(usd3)
	if err != nil {
		t.Fatalf("Cmp failed: %v", err)
	}
	if cmp != 0 {
		t.Errorf("100 == 100: got %d, want 0", cmp)
	}
}

func TestCmpAssetMismatch(t *testing.T) {
	usd := NewCurrency(USD())
	eur := NewCurrency(EUR())

	_, err := usd.Cmp(eur)
	if err == nil {
		t.Error("Expected error when comparing different assets")
	}
}

func TestCurrencyEqual(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("100.00")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("100.00")

	usd3 := NewCurrency(USD())
	usd3.SetAmountString("200.00")

	eur := NewCurrency(EUR())
	eur.SetAmountString("100.00")

	if !usd1.Equal(usd2) {
		t.Error("Same amount, same asset should be equal")
	}

	if usd1.Equal(usd3) {
		t.Error("Different amounts should not be equal")
	}

	if usd1.Equal(eur) {
		t.Error("Different assets should not be equal (even with same amount)")
	}
}

func TestAdd(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("100.50")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("50.25")

	err := usd1.Add(usd2)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if usd1.Amount() != "150.75" {
		t.Errorf("100.50 + 50.25: got %s, want 150.75", usd1.Amount())
	}
}

func TestAddNegative(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("100.00")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("-150.00")

	usd1.Add(usd2)

	if usd1.Amount() != "-50.00" {
		t.Errorf("100 + (-150): got %s, want -50.00", usd1.Amount())
	}
}

func TestAddAssetMismatch(t *testing.T) {
	usd := NewCurrency(USD())
	eur := NewCurrency(EUR())

	err := usd.Add(eur)
	if err == nil {
		t.Error("Expected error when adding different assets")
	}
}

func TestSub(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("100.50")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("50.25")

	err := usd1.Sub(usd2)
	if err != nil {
		t.Fatalf("Sub failed: %v", err)
	}

	if usd1.Amount() != "50.25" {
		t.Errorf("100.50 - 50.25: got %s, want 50.25", usd1.Amount())
	}
}

func TestSubResultNegative(t *testing.T) {
	usd1 := NewCurrency(USD())
	usd1.SetAmountString("50.00")

	usd2 := NewCurrency(USD())
	usd2.SetAmountString("100.00")

	usd1.Sub(usd2)

	if usd1.Amount() != "-50.00" {
		t.Errorf("50 - 100: got %s, want -50.00", usd1.Amount())
	}
}

func TestSubAssetMismatch(t *testing.T) {
	usd := NewCurrency(USD())
	eur := NewCurrency(EUR())

	err := usd.Sub(eur)
	if err == nil {
		t.Error("Expected error when subtracting different assets")
	}
}

func TestNeg(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"100.00", "-100.00"},
		{"-100.00", "100.00"},
		{"0.00", "0.00"},
		{"0.01", "-0.01"},
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		currency.SetAmountString(test.input)
		currency.Neg()

		if currency.Amount() != test.expected {
			t.Errorf("Neg(%s): got %s, want %s", test.input, currency.Amount(), test.expected)
		}
	}
}

func TestAbs(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"100.00", "100.00"},
		{"-100.00", "100.00"},
		{"0.00", "0.00"},
		{"-0.01", "0.01"},
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		currency.SetAmountString(test.input)
		currency.Abs()

		if currency.Amount() != test.expected {
			t.Errorf("Abs(%s): got %s, want %s", test.input, currency.Amount(), test.expected)
		}
	}
}

func TestCurrencyString(t *testing.T) {
	cases := []struct {
		asset    Asset
		amount   string
		expected string
	}{
		{USD(), "123.45", "$ 123.45"},
		{EUR(), "99.99", "€ 99.99"},
		{ZAR(), "1000.00", "R 1000.00"},
		{JPY(), "500", "¥ 500"}, // JPY has scale 0
		{USD(), "-50.00", "$ -50.00"},
	}

	for _, test := range cases {
		currency := NewCurrency(test.asset)
		currency.SetAmountString(test.amount)
		result := currency.String()

		if result != test.expected {
			t.Errorf("String() for %s %s: got %q, want %q", test.asset.Code(), test.amount, result, test.expected)
		}
	}
}

func TestAmountWithScaleZero(t *testing.T) {
	// JPY has scale 0, so no fractional part
	jpy := NewCurrency(JPY())
	jpy.SetAmountString("12345")

	if jpy.Amount() != "12345" {
		t.Errorf("JPY Amount: got %s, want 12345", jpy.Amount())
	}

	// Raw amount should equal the display amount for scale 0
	if jpy.RawAmount().Cmp(big.NewInt(12345)) != 0 {
		t.Errorf("JPY RawAmount: got %s, want 12345", jpy.RawAmount().String())
	}
}

func TestCode(t *testing.T) {
	cases := []struct {
		asset        Asset
		expectedCode string
	}{
		{USD(), "USD"},
		{EUR(), "EUR"},
		{ZAR(), "ZAR"},
		{JPY(), "JPY"},
		{CAD(), "CAD"},
	}
	for _, test := range cases {
		currency := NewCurrency(test.asset)
		code := currency.Code()
		if code != test.expectedCode {
			t.Errorf("For asset %s, expected code %s, got %s", test.asset.Code(), test.expectedCode, code)
		}
	}
}

func TestScale(t *testing.T) {
	cases := []struct {
		asset         Asset
		expectedScale uint8
	}{
		{USD(), 2},
		{EUR(), 2},
		{ZAR(), 2},
		{JPY(), 0},
		{CAD(), 2},
		{NewAsset("test", "999", 3, func(value string) string { return "test " + value }), 3},
	}
	for _, test := range cases {
		currency := NewCurrency(test.asset)
		scale := currency.Scale()
		if scale != test.expectedScale {
			t.Errorf("For asset %s, expected scale %d, got %d", test.asset.Code(), test.expectedScale, scale)
		}
	}
}

func TestNumericCode(t *testing.T) {
	cases := []struct {
		asset               Asset
		expectedNumericCode string
	}{
		{USD(), "840"},
		{EUR(), "978"},
		{ZAR(), "710"},
		{JPY(), "392"},
		{CAD(), "124"},
		{NewAsset("test", "999", 3, func(value string) string { return "test " + value }), "999"},
	}
	for _, test := range cases {
		currency := NewCurrency(test.asset)
		numericCode := currency.NumericCode()
		if numericCode != test.expectedNumericCode {
			t.Errorf("For asset %s, expected numeric code %s, got %s", test.asset.Code(), test.expectedNumericCode, numericCode)
		}
	}
}

func TestAmount(t *testing.T) {
	// trial to verify that new Currency starts with amount zero
	usd := NewCurrency(USD())
	eur := NewCurrency(EUR())
	zar := NewCurrency(ZAR())
	jpy := NewCurrency(JPY())
	cad := NewCurrency(CAD())

	cases := []struct {
		asset       *Currency
		expectedAmt *big.Int
	}{
		{usd, big.NewInt(0)},
		{eur, big.NewInt(0)},
		{zar, big.NewInt(0)},
		{jpy, big.NewInt(0)},
		{cad, big.NewInt(0)},
	}

	for _, test := range cases {
		amount := test.asset.amount
		if amount.Cmp(test.expectedAmt) != 0 {
			t.Errorf("For asset %s, expected amount %s, got %s", test.asset.Code(), test.expectedAmt.String(), amount.String())
		}
	}

	// Test non-zero amounts with SetAmountString
	testCurrency := NewCurrency(USD())

	c := *testCurrency
	if err := c.SetAmountString("123.45"); err != nil {
		t.Fatalf("SetAmountString(123.45) failed: %v", err)
	}
	if c.amount.Cmp(big.NewInt(12345)) != 0 {
		t.Errorf("SetAmountString(123.45): expected 12345, got %s", c.amount.String())
	}

	c = *testCurrency
	if err := c.SetAmountString("-67.89"); err != nil {
		t.Fatalf("SetAmountString(-67.89) failed: %v", err)
	}
	if c.amount.Cmp(big.NewInt(-6789)) != 0 {
		t.Errorf("SetAmountString(-67.89): expected -6789, got %s", c.amount.String())
	}

}

func TestCurrencyFactor(t *testing.T) {
	cases := []struct {
		asset        Asset
		expectedFact *big.Int
	}{
		{USD(), big.NewInt(100)}, // scale 2
		{EUR(), big.NewInt(100)}, // scale 2
		{ZAR(), big.NewInt(100)}, // scale 2
		{JPY(), big.NewInt(1)},   // scale 0
		{CAD(), big.NewInt(100)}, // scale 2
	}
	for _, test := range cases {
		currency := NewCurrency(test.asset)
		factor := currency.Factor()
		if factor.Cmp(test.expectedFact) != 0 {
			t.Errorf("For asset %s, expected factor %s, got %s", test.asset.Code(), test.expectedFact.String(), factor.String())
		}
	}
}

func TestSetAmountString(t *testing.T) {
	positiveBigInt := new(big.Int).Lsh(big.NewInt(1), 200) // 2^200, a very large number
	negativeBigInt := new(big.Int).Neg(positiveBigInt)

	factor := NewCurrency(USD()).Factor()

	cases := []struct {
		input       string
		expectError bool
		expected    big.Int
	}{
		{"0", false, *big.NewInt(0)},
		{"70", false, *big.NewInt(7000)},
		{"-80", false, *big.NewInt(-8000)},
		{"92233720368547758", false, *big.NewInt(9223372036854775800)},
		{"-92233720368547758", false, *big.NewInt(-9223372036854775800)},
		{positiveBigInt.String(), false, *new(big.Int).Mul(positiveBigInt, factor)},
		{negativeBigInt.String(), false, *new(big.Int).Mul(negativeBigInt, factor)},

		{"12.34", false, *big.NewInt(1234)},
		{"-56.78", false, *big.NewInt(-5678)},
		{"3.14159", false, *big.NewInt(314)},
		{"-3.14159", false, *big.NewInt(-314)},
		{"0.1", false, *big.NewInt(10)},
		{"-0.1", false, *big.NewInt(-10)},
		{"0.12", false, *big.NewInt(12)},
		{"-0.12", false, *big.NewInt(-12)},
		{"0.123", false, *big.NewInt(12)},   // Rounds down
		{"-0.123", false, *big.NewInt(-12)}, // Rounds down
		{"1.05", false, *big.NewInt(105)},
		{"-1.05", false, *big.NewInt(-105)},
		{"19.3", false, *big.NewInt(1930)},
		{"-19.3", false, *big.NewInt(-1930)},
		{"0.001", false, *big.NewInt(0)},
		{"-0.001", false, *big.NewInt(0)},

		{"-0", false, *big.NewInt(0)},

		{"+0.0", false, *big.NewInt(0)},
		{"-0.0", false, *big.NewInt(0)},
		{"+0.01", false, *big.NewInt(1)},
		{"-0.01", false, *big.NewInt(-1)},

		{"19.9999999", false, *big.NewInt(1999)},
		{"-19.9999999", false, *big.NewInt(-1999)},
		{"999999999999.9999", false, *big.NewInt(99999999999999)},
		{"-999999999999.9999", false, *big.NewInt(-99999999999999)},

		// valid comma-separated
		{"1,234", false, *big.NewInt(123400)},
		{"1,234.56", false, *big.NewInt(123456)},
		{"12,345.67", false, *big.NewInt(1234567)},
		{"1,234,567.89", false, *big.NewInt(123456789)},
		{"-1,234.56", false, *big.NewInt(-123456)},
		{"+1,234.56", false, *big.NewInt(123456)},

		// invalid comma placement
		{",123", true, *big.NewInt(0)},          // leading comma
		{"1,,234", true, *big.NewInt(0)},         // consecutive commas
		{"1,23", true, *big.NewInt(0)},           // wrong grouping
		{"1234,567", true, *big.NewInt(0)},       // first group too large
		{"1,234,56", true, *big.NewInt(0)},       // last group too small
		{"1,,,,2.34", true, *big.NewInt(0)},      // multiple consecutive commas
		{"1.234,56", true, *big.NewInt(0)},       // comma in fractional part
		{",", true, *big.NewInt(0)},              // just a comma
		{"-,123", true, *big.NewInt(0)},          // sign then comma
		{"1,234,", true, *big.NewInt(0)},         // trailing comma

		// invalid strings
		{"", true, *big.NewInt(0)},
		{"invalid", true, *big.NewInt(0)},
		{".", true, *big.NewInt(0)},        // just a dot
		{"123.", true, *big.NewInt(0)},     // trailing dot, empty fractional
		{".123", true, *big.NewInt(0)},     // leading dot, empty whole
		{"1.2.3", true, *big.NewInt(0)},    // multiple dots
		{"12..34", true, *big.NewInt(0)},   // consecutive dots
		{" 123.45", true, *big.NewInt(0)},  // leading whitespace
		{"123.45 ", true, *big.NewInt(0)},  // trailing whitespace
		{"1 234.56", true, *big.NewInt(0)}, // space in number
		{"1e10", true, *big.NewInt(0)},     // scientific notation
		{"1.5e-3", true, *big.NewInt(0)},   // scientific notation
		{"$100", true, *big.NewInt(0)},     // currency symbol
		{"100USD", true, *big.NewInt(0)},   // letters
		{"--100", true, *big.NewInt(0)},    // double negative
		{"++100", true, *big.NewInt(0)},    // double positive
		{"+-100", true, *big.NewInt(0)},    // mixed signs
		{"abc.12", true, *big.NewInt(0)},   // invalid whole part (letters)
		{"1a2.34", true, *big.NewInt(0)},   // invalid whole part (mixed)
		{"12.ab", true, *big.NewInt(0)},    // invalid fractional part (letters)
		{"12.3a", true, *big.NewInt(0)},    // invalid fractional part (mixed)
	}

	for _, test := range cases {
		currency := NewCurrency(USD())
		err := currency.SetAmountString(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect error for input %q, but got: %v", test.input, err)
			} else if currency.amount.Cmp(&test.expected) != 0 {
				t.Errorf("For input %q, expected amount %s, got %s", test.input, test.expected.String(), currency.amount.String())
			}
		}
	}
}

func BenchmarkSetAmountString(b *testing.B) {
	currency := NewCurrency(USD())
	input := "12345.67"

	for b.Loop() {
		err := currency.SetAmountString(input)
		if err != nil {
			b.Errorf("SetAmountString failed: %v", err)
		}
	}
}

func TestCurrencyToProtoGeoV1(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		asset       Asset
		countryCode string
	}{
		{
			name:        "USD with country code",
			amount:      "100.00",
			asset:       USD(),
			countryCode: "US",
		},
		{
			name:        "EUR without country code",
			amount:      "50.50",
			asset:       EUR(),
			countryCode: "",
		},
		{
			name:        "JPY with zero scale",
			amount:      "1000",
			asset:       JPY(),
			countryCode: "JP",
		},
		{
			name:        "negative amount",
			amount:      "-25.99",
			asset:       USD(),
			countryCode: "US",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency := NewCurrency(tt.asset)
			currency.SetAmountString(tt.amount)

			pb := currency.ToProtoGeoV1(tt.countryCode)

			if pb == nil {
				t.Fatal("ToProtoGeoV1() returned nil")
			}
			if pb.CountryCode != tt.countryCode {
				t.Errorf("ToProtoGeoV1().CountryCode = %v, want %v", pb.CountryCode, tt.countryCode)
			}
			if pb.Asset == nil {
				t.Fatal("ToProtoGeoV1().Asset is nil")
			}
			if pb.Asset.Code != tt.asset.Code() {
				t.Errorf("ToProtoGeoV1().Asset.Code = %v, want %v", pb.Asset.Code, tt.asset.Code())
			}
			// Verify amount is the raw amount string
			expectedRaw := currency.RawAmount().String()
			if pb.Amount != expectedRaw {
				t.Errorf("ToProtoGeoV1().Amount = %v, want %v", pb.Amount, expectedRaw)
			}
		})
	}
}

func TestCurrencyFromProtoGeoV1(t *testing.T) {
	tests := []struct {
		name       string
		pb         *geopbv1.Currency
		wantAmount string
		wantAsset  string
		wantErr    bool
	}{
		{
			name: "valid USD currency",
			pb: &geopbv1.Currency{
				Amount: "10000",
				Asset: &geopbv1.Asset{
					Code:    "USD",
					Numeric: "840",
					Scale:   2,
				},
				CountryCode: "US",
			},
			wantAmount: "100.00",
			wantAsset:  "USD",
			wantErr:    false,
		},
		{
			name: "valid JPY currency",
			pb: &geopbv1.Currency{
				Amount: "1000",
				Asset: &geopbv1.Asset{
					Code:    "JPY",
					Numeric: "392",
					Scale:   0,
				},
				CountryCode: "JP",
			},
			wantAmount: "1000",
			wantAsset:  "JPY",
			wantErr:    false,
		},
		{
			name: "negative amount",
			pb: &geopbv1.Currency{
				Amount: "-2599",
				Asset: &geopbv1.Asset{
					Code:    "USD",
					Numeric: "840",
					Scale:   2,
				},
				CountryCode: "US",
			},
			wantAmount: "-25.99",
			wantAsset:  "USD",
			wantErr:    false,
		},
		{
			name:       "nil proto returns error",
			pb:         nil,
			wantAmount: "",
			wantAsset:  "",
			wantErr:    true,
		},
		{
			name: "unsupported asset returns error",
			pb: &geopbv1.Currency{
				Amount: "100",
				Asset: &geopbv1.Asset{
					Code:    "XXX",
					Numeric: "999",
					Scale:   2,
				},
				CountryCode: "XX",
			},
			wantAmount: "",
			wantAsset:  "",
			wantErr:    true,
		},
		{
			name: "invalid amount returns error",
			pb: &geopbv1.Currency{
				Amount: "not-a-number",
				Asset: &geopbv1.Asset{
					Code:    "USD",
					Numeric: "840",
					Scale:   2,
				},
				CountryCode: "US",
			},
			wantAmount: "",
			wantAsset:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, err := CurrencyFromProtoGeoV1(tt.pb)
			if (err != nil) != tt.wantErr {
				t.Errorf("CurrencyFromProtoGeoV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if currency.Amount() != tt.wantAmount {
				t.Errorf("CurrencyFromProtoGeoV1().Amount() = %v, want %v", currency.Amount(), tt.wantAmount)
			}
			if currency.Code() != tt.wantAsset {
				t.Errorf("CurrencyFromProtoGeoV1().Code() = %v, want %v", currency.Code(), tt.wantAsset)
			}
		})
	}
}
