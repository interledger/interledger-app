package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/rafiki"
)

func TestIncomingPaymentURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		walletAddrURL  string
		id             string
		want           string
	}{
		{
			name:          "standard wallet address URL",
			walletAddrURL: "https://wallet.example.org/alice",
			id:            "abc-123",
			want:          "https://wallet.example.org/alice/incoming-payments/abc-123",
		},
		{
			name:          "trailing slash on wallet address URL is normalised",
			walletAddrURL: "https://wallet.example.org/alice/",
			id:            "abc-123",
			want:          "https://wallet.example.org/alice/incoming-payments/abc-123",
		},
		{
			name:          "empty wallet address URL falls back to bare id",
			walletAddrURL: "",
			id:            "abc-123",
			want:          "abc-123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, incomingPaymentURL(tc.walletAddrURL, tc.id))
		})
	}
}

func TestParseRafikiTime(t *testing.T) {
	t.Parallel()

	want := time.Date(2026, 5, 12, 9, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input string
		ok    bool
		out   time.Time
	}{
		{"empty string", "", false, time.Time{}},
		{"unparseable", "not-a-date", false, time.Time{}},
		{"rfc3339 UTC", "2026-05-12T09:30:00Z", true, want},
		{"rfc3339 with offset", "2026-05-12T11:30:00+02:00", true, want},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseRafikiTime(tc.input)
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.True(t, got.Equal(tc.out), "expected %s, got %s", tc.out, got)
			}
		})
	}
}

func TestTransformIncomingPaymentRequest(t *testing.T) {
	t.Parallel()

	amt := currency.FromUInt64(1250, currency.EUR)
	ip := &rafiki.IncomingPayment{
		ID:              "ip_42",
		WalletAddressID: "wa_1",
		State:           rafiki.IncomingPaymentStatePending,
		ExpiresAt:       "2026-05-13T00:00:00Z",
		CreatedAt:       "2026-05-12T09:30:00Z",
		IncomingAmount:  &amt,
		Metadata:        map[string]interface{}{"description": "Rent"},
	}
	wa := &rafiki.WalletAddress{
		ID:         "wa_1",
		AssetCode:  "EUR",
		AssetScale: 2,
		URL:        "https://wallet.example.org/alice",
	}

	got := transformIncomingPaymentRequest(ip, wa)

	assert.Equal(t, "ip_42", got.GetId())
	assert.Equal(t, "https://wallet.example.org/alice/incoming-payments/ip_42", got.GetUrl())
	assert.Equal(t, "Rent", got.GetDescription())
	assert.Equal(t, string(rafiki.IncomingPaymentStatePending), got.GetState())
	assert.NotNil(t, got.GetIncomingAmount())
	assert.Equal(t, int64(1250), got.GetIncomingAmount().GetAmount())
	assert.NotNil(t, got.GetExpiresAt())
	assert.Equal(t, int64(1778630400), got.GetExpiresAt().GetSeconds())
	assert.NotNil(t, got.GetCreatedAt())
	assert.Equal(t, int64(1778578200), got.GetCreatedAt().GetSeconds())
}

func TestTransformIncomingPaymentRequest_MissingOptionals(t *testing.T) {
	t.Parallel()

	ip := &rafiki.IncomingPayment{
		ID:              "ip_43",
		WalletAddressID: "wa_1",
		State:           rafiki.IncomingPaymentStatePending,
	}
	wa := &rafiki.WalletAddress{
		ID:  "wa_1",
		URL: "https://wallet.example.org/alice",
	}

	got := transformIncomingPaymentRequest(ip, wa)

	assert.Equal(t, "ip_43", got.GetId())
	assert.Empty(t, got.GetDescription())
	assert.Nil(t, got.GetIncomingAmount())
	assert.Nil(t, got.GetExpiresAt())
	assert.Nil(t, got.GetCreatedAt())
}

func TestMetadataDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		md   map[string]interface{}
		want string
	}{
		{"nil map", nil, ""},
		{"empty map", map[string]interface{}{}, ""},
		{"description present", map[string]interface{}{"description": "Rent"}, "Rent"},
		{"description wrong type", map[string]interface{}{"description": 42}, ""},
		{"description with unrelated keys", map[string]interface{}{"description": "Lunch", "other": true}, "Lunch"},
		{"missing key", map[string]interface{}{"other": "value"}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, metadataDescription(tc.md))
		})
	}
}
