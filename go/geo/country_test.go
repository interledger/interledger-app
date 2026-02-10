package geo

import (
	"testing"
)

func TestParseCountry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Country
	}{
		{
			name:  "valid alpha2 uppercase",
			input: "US",
			want:  US,
		},
		{
			name:  "valid alpha2 lowercase",
			input: "us",
			want:  US,
		},
		{
			name:  "valid alpha2 mixed case",
			input: "Us",
			want:  US,
		},
		{
			name:  "valid alpha2 with whitespace",
			input: "  ZA  ",
			want:  ZA,
		},
		{
			name:  "valid numeric code",
			input: "840",
			want:  US,
		},
		{
			name:  "valid numeric code for ZA",
			input: "710",
			want:  ZA,
		},
		{
			name:  "invalid code defaults to US",
			input: "INVALID",
			want:  US,
		},
		{
			name:  "empty string defaults to US",
			input: "",
			want:  US,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCountry(tt.input)
			if got != tt.want {
				t.Errorf("ParseCountry(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCountryNumeric(t *testing.T) {
	tests := []struct {
		name      string
		country   Country
		want      string
		wantError bool
	}{
		{
			name:      "US numeric",
			country:   US,
			want:      "840",
			wantError: false,
		},
		{
			name:      "ZA numeric",
			country:   ZA,
			want:      "710",
			wantError: false,
		},
		{
			name:      "GB numeric",
			country:   GB,
			want:      "826",
			wantError: false,
		},
		{
			name:      "invalid country",
			country:   Country("XX"),
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.country.Numeric()
			if (err != nil) != tt.wantError {
				t.Errorf("Country.Numeric() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("Country.Numeric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountryValid(t *testing.T) {
	tests := []struct {
		name    string
		country Country
		want    bool
	}{
		{
			name:    "US is valid",
			country: US,
			want:    true,
		},
		{
			name:    "ZA is valid",
			country: ZA,
			want:    true,
		},
		{
			name:    "invalid country",
			country: Country("XX"),
			want:    false,
		},
		{
			name:    "empty country",
			country: Country(""),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.country.Valid(); got != tt.want {
				t.Errorf("Country.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountryString(t *testing.T) {
	tests := []struct {
		name    string
		country Country
		want    string
	}{
		{
			name:    "US string",
			country: US,
			want:    "US",
		},
		{
			name:    "ZA string",
			country: ZA,
			want:    "ZA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.country.String(); got != tt.want {
				t.Errorf("Country.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountryIsSupported(t *testing.T) {
	// Note: This test assumes some countries are marked as supported in the details map.
	// We test both the method returning true (if any supported) and false cases.
	tests := []struct {
		name    string
		country Country
	}{
		{
			name:    "invalid country is not supported",
			country: Country("XX"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Invalid countries should return false
			if got := tt.country.IsSupported(); got != false {
				t.Errorf("Country.IsSupported() for invalid country = %v, want false", got)
			}
		})
	}

	// Test that valid countries return a boolean based on their Supported field
	// We just verify it doesn't panic and returns a boolean
	_ = US.IsSupported()
	_ = ZA.IsSupported()
}

func TestIsEUCountry(t *testing.T) {
	tests := []struct {
		name    string
		country Country
		want    bool
	}{
		{
			name:    "Germany is EU",
			country: DE,
			want:    true,
		},
		{
			name:    "France is EU",
			country: FR,
			want:    true,
		},
		{
			name:    "Netherlands is EU",
			country: NL,
			want:    true,
		},
		{
			name:    "US is not EU",
			country: US,
			want:    false,
		},
		{
			name:    "ZA is not EU",
			country: ZA,
			want:    false,
		},
		{
			name:    "GB is not EU",
			country: GB,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEUCountry(tt.country); got != tt.want {
				t.Errorf("IsEUCountry(%v) = %v, want %v", tt.country, got, tt.want)
			}
		})
	}
}

func TestEUCountries(t *testing.T) {
	countries := EUCountries()

	// Check that we have the expected number of EU countries (27 as of 2024)
	if len(countries) != 27 {
		t.Errorf("EUCountries() returned %d countries, want 27", len(countries))
	}

	// Verify all returned countries are actually EU countries
	for _, c := range countries {
		if !IsEUCountry(c) {
			t.Errorf("EUCountries() returned %v which is not an EU country", c)
		}
	}

	// Verify some expected EU countries are in the list
	expectedEU := []Country{DE, FR, IT, ES, NL, BE, AT, PL}
	countrySet := make(map[Country]bool)
	for _, c := range countries {
		countrySet[c] = true
	}

	for _, expected := range expectedEU {
		if !countrySet[expected] {
			t.Errorf("EUCountries() missing expected country %v", expected)
		}
	}
}

func TestGetCountryDetail(t *testing.T) {
	tests := []struct {
		name     string
		country  Country
		wantName string
		wantOk   bool
	}{
		{
			name:     "US detail",
			country:  US,
			wantName: "United States of America",
			wantOk:   true,
		},
		{
			name:     "ZA detail",
			country:  ZA,
			wantName: "South Africa",
			wantOk:   true,
		},
		{
			name:     "GB detail",
			country:  GB,
			wantName: "United Kingdom of Great Britain and Northern Ireland",
			wantOk:   true,
		},
		{
			name:     "invalid country",
			country:  Country("XX"),
			wantName: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail, ok := GetCountryDetail(tt.country)
			if ok != tt.wantOk {
				t.Errorf("GetCountryDetail(%v) ok = %v, want %v", tt.country, ok, tt.wantOk)
				return
			}
			if ok && detail.Name != tt.wantName {
				t.Errorf("GetCountryDetail(%v).Name = %v, want %v", tt.country, detail.Name, tt.wantName)
			}
		})
	}
}

func TestAllCountries(t *testing.T) {
	countries := AllCountries()

	// Should return a non-empty slice
	if len(countries) == 0 {
		t.Error("AllCountries() returned empty slice")
	}

	// All returned countries should be valid
	for _, c := range countries {
		if !c.Valid() {
			t.Errorf("AllCountries() returned invalid country %v", c)
		}
	}

	// Check that some expected countries are present
	expected := []Country{US, GB, ZA, DE, FR, CA, AU}
	countrySet := make(map[Country]bool)
	for _, c := range countries {
		countrySet[c] = true
	}

	for _, e := range expected {
		if !countrySet[e] {
			t.Errorf("AllCountries() missing expected country %v", e)
		}
	}
}

func TestGetStates(t *testing.T) {
	tests := []struct {
		name    string
		country Country
		wantOk  bool
	}{
		{
			name:    "US has states",
			country: US,
			wantOk:  true,
		},
		{
			name:    "GB has states",
			country: GB,
			wantOk:  true,
		},
		{
			name:    "invalid country has no states",
			country: Country("XX"),
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			states, ok := GetStates(tt.country)
			if ok != tt.wantOk {
				t.Errorf("GetStates(%v) ok = %v, want %v", tt.country, ok, tt.wantOk)
				return
			}
			if ok && len(states) == 0 {
				t.Errorf("GetStates(%v) returned empty map", tt.country)
			}
		})
	}

	// Verify that modifications to returned map don't affect internal state
	states1, _ := GetStates(US)
	states1["TEST"] = "Test State"

	states2, _ := GetStates(US)
	if _, exists := states2["TEST"]; exists {
		t.Error("GetStates() returned map that allows modification of internal state")
	}
}

func TestGetStateName(t *testing.T) {
	tests := []struct {
		name      string
		country   Country
		stateCode string
		wantName  string
		wantOk    bool
	}{
		{
			name:      "US California",
			country:   US,
			stateCode: "CA",
			wantName:  "CALIFORNIA",
			wantOk:    true,
		},
		{
			name:      "US New York",
			country:   US,
			stateCode: "NY",
			wantName:  "NEW YORK",
			wantOk:    true,
		},
		{
			name:      "US Texas",
			country:   US,
			stateCode: "TX",
			wantName:  "TEXAS",
			wantOk:    true,
		},
		{
			name:      "invalid state code",
			country:   US,
			stateCode: "XX",
			wantName:  "",
			wantOk:    false,
		},
		{
			name:      "invalid country",
			country:   Country("XX"),
			stateCode: "CA",
			wantName:  "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, ok := GetStateName(tt.country, tt.stateCode)
			if ok != tt.wantOk {
				t.Errorf("GetStateName(%v, %v) ok = %v, want %v", tt.country, tt.stateCode, ok, tt.wantOk)
				return
			}
			if name != tt.wantName {
				t.Errorf("GetStateName(%v, %v) = %v, want %v", tt.country, tt.stateCode, name, tt.wantName)
			}
		})
	}
}

func TestToAlpha3(t *testing.T) {
	tests := []struct {
		name   string
		alpha2 string
		want   string
	}{
		{
			name:   "US to USA",
			alpha2: "US",
			want:   "USA",
		},
		{
			name:   "lowercase us to USA",
			alpha2: "us",
			want:   "USA",
		},
		{
			name:   "GB to GBR",
			alpha2: "GB",
			want:   "GBR",
		},
		{
			name:   "ZA to ZAF",
			alpha2: "ZA",
			want:   "ZAF",
		},
		{
			name:   "DE to DEU",
			alpha2: "DE",
			want:   "DEU",
		},
		{
			name:   "FR to FRA",
			alpha2: "FR",
			want:   "FRA",
		},
		{
			name:   "CA to CAN",
			alpha2: "CA",
			want:   "CAN",
		},
		{
			name:   "AU to AUS",
			alpha2: "AU",
			want:   "AUS",
		},
		{
			name:   "JP to JPN",
			alpha2: "JP",
			want:   "JPN",
		},
		{
			name:   "invalid alpha2 returns empty",
			alpha2: "XX",
			want:   "",
		},
		{
			name:   "empty string returns empty",
			alpha2: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToAlpha3(tt.alpha2); got != tt.want {
				t.Errorf("ToAlpha3(%v) = %v, want %v", tt.alpha2, got, tt.want)
			}
		})
	}
}
