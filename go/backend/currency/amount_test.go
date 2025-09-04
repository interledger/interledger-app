package currency_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/currency"
)

func TestCurrency(t *testing.T) {
	t.Parallel()
		
	cases := []struct {
		cc    string
		valid bool
		scale int
	}{
		{
			cc:    "USD",
			valid: true,
			scale: 2,
		},
		{
			cc:    "ZAR",
			valid: true,
			scale: 2,
		},
		{
			cc:    "THB",
			valid: false,
			scale: 2,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("currency=%s", tc.cc), func(t *testing.T) {
			c := currency.ParseCurrency(tc.cc)
			assert.Equal(t, tc.valid, c.Valid())
			assert.Equal(t, tc.scale, c.Scale())
		})
	}
}

func TestAmount_Float64(t *testing.T) {
	t.Parallel()
		
	cases := []struct {
		name string
		in   currency.Amount
		out  float64
	}{
		{
			name: "currency and scale",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
				Scale:    2,
			},
			out: 10.00,
		},
		{
			name: "scale omitted",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
			},
			out: 10.00,
		},
		{
			name: "currency and scale omitted",
			in: currency.Amount{
				Value: 1000,
			},
			out: 10.00, // Default to 2 decimal points
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.out, tc.in.Float64())
		})
	}
}

func TestAmount_FormatAmount(t *testing.T) {
	t.Parallel()
		
	cases := []struct {
		name string
		in   currency.Amount
		out  string
	}{
		{
			name: "currency and scale",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
				Scale:    2,
			},
			out: "10.00",
		},
		{
			name: "scale omitted",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
			},
			out: "10.00",
		},
		{
			name: "currency and scale omitted",
			in: currency.Amount{
				Value: 1000,
			},
			out: "10.00", // Default to 2 decimal points
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.out, tc.in.FormatAmount())
		})
	}
}

func TestAmount_Format(t *testing.T) {
	t.Parallel()
	
	cases := []struct {
		name string
		in   currency.Amount
		out  string
	}{
		{
			name: "currency and scale",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
				Scale:    2,
			},
			out: "$ 10.00",
		},
		{
			name: "scale omitted",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.USD,
			},
			out: "$ 10.00",
		},
		{
			name: "currency without a format",
			in: currency.Amount{
				Value:    1000,
				Currency: currency.ParseCurrency("THB"),
			},
			out: "10.00 THB", // Default to 2 decimal points
		},
		{
			name: "currency without a format scale provided",
			in: currency.Amount{
				Value:    1000,
				Scale:    8,
				Currency: currency.ParseCurrency("BTC"),
			},
			out: "0.00001000 BTC",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.out, tc.in.Format())
		})
	}
}

func TestFromFloat(t *testing.T) {
	t.Parallel()
		
	cases := []struct {
		name    string
		in      float64
		inCC    string
		format  string
		float64 float64
	}{
		{
			name:    "10 USD",
			in:      10.00,
			inCC:    "USD",
			format:  "$ 10.00",
			float64: 10.00,
		},
		{
			name:    "too many decimals",
			in:      10.0012,
			inCC:    "USD",
			format:  "$ 10.00",
			float64: 10.00,
		},
		{
			name:    "unknown currency",
			in:      10.00,
			inCC:    "THB",
			format:  "10.00 THB",
			float64: 10.00,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := currency.FromFloat64(tc.in, currency.ParseCurrency(tc.inCC))

			assert.Equal(t, tc.float64, a.Float64())
			assert.Equal(t, tc.format, a.Format())
		})
	}
}
