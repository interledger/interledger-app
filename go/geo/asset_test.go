package geo

import (
	"math/big"
	"testing"

	geopbv1 "gitlab.com/fynbos/proto/geo/v1"
)

func TestNewAsset(t *testing.T) {
	a := NewAsset("TST", "999", 3, func(value string) string { return "T " + value })

	if a.Code() != "TST" {
		t.Errorf("Code: got %s, want TST", a.Code())
	}
	if a.NumericCode() != "999" {
		t.Errorf("NumericCode: got %s, want 999", a.NumericCode())
	}
	if a.Scale() != 3 {
		t.Errorf("Scale: got %d, want 3", a.Scale())
	}
	if a.Format("100") != "T 100" {
		t.Errorf("Format: got %s, want T 100", a.Format("100"))
	}
}

func TestNewAssetNilFormat(t *testing.T) {
	// Should not panic when formatFunc is nil
	a := NewAsset("TST", "999", 2, nil)

	if a.Format("100") != "100" {
		t.Errorf("Format with nil func: got %s, want 100", a.Format("100"))
	}
}

func TestAssetFactor(t *testing.T) {
	tests := []struct {
		name  string
		asset Asset
		want  *big.Int
	}{
		{
			name:  "USD factor",
			asset: USD(),
			want:  big.NewInt(100),
		},
		{
			name:  "JPY factor",
			asset: JPY(),
			want:  big.NewInt(1),
		},
		{
			name:  "CAD factor",
			asset: CAD(),
			want:  big.NewInt(100),
		},
		{
			name:  "test asset with scale 3",
			asset: NewAsset("TST", "999", 3, func(value string) string { return "T" + value }),
			want:  big.NewInt(1000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := tt.asset
			if got := asset.Factor(); got.Cmp(tt.want) != 0 {
				t.Errorf("Asset.factor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssetEqual(t *testing.T) {
	usd1 := USD()
	usd2 := NewAsset("USD", "840", 2, func(value string) string { return "$ " + value })
	eur := EUR()

	if !usd1.Equal(usd2) {
		t.Error("Same code should be equal")
	}

	if usd1.Equal(eur) {
		t.Error("Different codes should not be equal")
	}
}

func TestAssetString(t *testing.T) {
	s := USD().String()
	if s != "Asset{code=USD, numeric=840, scale=2}" {
		t.Errorf("String: got %s", s)
	}
}

func TestPredefinedAssets(t *testing.T) {
	cases := []struct {
		asset   Asset
		code    string
		numeric string
		scale   uint8
	}{
		{USD(), "USD", "840", 2},
		{EUR(), "EUR", "978", 2},
		{ZAR(), "ZAR", "710", 2},
		{CAD(), "CAD", "124", 2},
		{JPY(), "JPY", "392", 0},
		{CAD(), "CAD", "124", 2},
	}

	for _, tc := range cases {
		if tc.asset.Code() != tc.code {
			t.Errorf("%s Code: got %s, want %s", tc.code, tc.asset.Code(), tc.code)
		}
		if tc.asset.NumericCode() != tc.numeric {
			t.Errorf("%s NumericCode: got %s, want %s", tc.code, tc.asset.NumericCode(), tc.numeric)
		}
		if tc.asset.Scale() != tc.scale {
			t.Errorf("%s Scale: got %d, want %d", tc.code, tc.asset.Scale(), tc.scale)
		}
	}
}

func TestGetAsset(t *testing.T) {
	// Test existing asset
	asset, ok := GetAsset("USD")
	if !ok {
		t.Error("GetAsset(USD) should return true")
	}
	if asset.Code() != "USD" {
		t.Errorf("GetAsset(USD) code: got %s, want USD", asset.Code())
	}

	// Test non-existing asset
	_, ok = GetAsset("XXX")
	if ok {
		t.Error("GetAsset(XXX) should return false")
	}

	// Test case sensitivity
	_, ok = GetAsset("usd")
	if ok {
		t.Error("GetAsset(usd) should return false (case-sensitive)")
	}
}

func TestIsSupported(t *testing.T) {
	cases := []struct {
		code     string
		expected bool
	}{
		{"USD", true},
		{"EUR", true},
		{"ZAR", true},
		{"CAD", true},
		{"JPY", true},
		{"XXX", false},
		{"usd", false}, // case-sensitive
		{"", false},
	}

	for _, tc := range cases {
		if got := IsSupported(tc.code); got != tc.expected {
			t.Errorf("IsSupported(%s): got %v, want %v", tc.code, got, tc.expected)
		}
	}
}

func TestAllAssets(t *testing.T) {
	all := AllAssets()

	// Check we have all expected assets
	if len(all) != 5 {
		t.Errorf("AllAssets length: got %d, want 5", len(all))
	}

	// Verify all expected codes are present
	codes := make(map[string]bool)
	for _, a := range all {
		codes[a.Code()] = true
	}

	expected := []string{"USD", "EUR", "ZAR", "CAD", "JPY"}
	for _, code := range expected {
		if !codes[code] {
			t.Errorf("AllAssets missing %s", code)
		}
	}
}

func TestAssetToProtoGeoV1(t *testing.T) {
	tests := []struct {
		name  string
		asset Asset
	}{
		{
			name:  "USD to proto",
			asset: USD(),
		},
		{
			name:  "EUR to proto",
			asset: EUR(),
		},
		{
			name:  "JPY to proto",
			asset: JPY(),
		},
		{
			name:  "custom asset to proto",
			asset: NewAsset("TST", "999", 3, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := tt.asset.ToProtoGeoV1()

			if pb == nil {
				t.Fatal("ToProtoGeoV1() returned nil")
			}
			if pb.Code != tt.asset.Code() {
				t.Errorf("ToProtoGeoV1().Code = %v, want %v", pb.Code, tt.asset.Code())
			}
			if pb.Numeric != tt.asset.NumericCode() {
				t.Errorf("ToProtoGeoV1().Numeric = %v, want %v", pb.Numeric, tt.asset.NumericCode())
			}
			if pb.Scale != uint32(tt.asset.Scale()) {
				t.Errorf("ToProtoGeoV1().Scale = %v, want %v", pb.Scale, tt.asset.Scale())
			}
		})
	}
}

func TestAssetFromProtoGeoV1(t *testing.T) {
	tests := []struct {
		name     string
		pb       *geopbv1.Asset
		wantCode string
		wantOk   bool
	}{
		{
			name: "USD from proto",
			pb: &geopbv1.Asset{
				Code:    "USD",
				Numeric: "840",
				Scale:   2,
			},
			wantCode: "USD",
			wantOk:   true,
		},
		{
			name: "EUR from proto",
			pb: &geopbv1.Asset{
				Code:    "EUR",
				Numeric: "978",
				Scale:   2,
			},
			wantCode: "EUR",
			wantOk:   true,
		},
		{
			name: "JPY from proto",
			pb: &geopbv1.Asset{
				Code:    "JPY",
				Numeric: "392",
				Scale:   0,
			},
			wantCode: "JPY",
			wantOk:   true,
		},
		{
			name:     "nil proto returns false",
			pb:       nil,
			wantCode: "",
			wantOk:   false,
		},
		{
			name: "unsupported asset returns false",
			pb: &geopbv1.Asset{
				Code:    "XXX",
				Numeric: "999",
				Scale:   2,
			},
			wantCode: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, ok := AssetFromProtoGeoV1(tt.pb)
			if ok != tt.wantOk {
				t.Errorf("AssetFromProtoGeoV1() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if ok && asset.Code() != tt.wantCode {
				t.Errorf("AssetFromProtoGeoV1().Code() = %v, want %v", asset.Code(), tt.wantCode)
			}
		})
	}
}
