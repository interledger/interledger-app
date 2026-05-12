package ops

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIncomingPaymentMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want incomingPaymentMetadata
	}{
		{
			name: "empty raw message",
			raw:  "",
			want: incomingPaymentMetadata{},
		},
		{
			name: "explicit null",
			raw:  "null",
			want: incomingPaymentMetadata{},
		},
		{
			name: "description only",
			raw:  `{"description":"Pizza night"}`,
			want: incomingPaymentMetadata{Description: "Pizza night"},
		},
		{
			name: "description with unknown keys",
			raw:  `{"description":"Rent","parentPaymentId":"pp_123","unknownField":42}`,
			want: incomingPaymentMetadata{Description: "Rent"},
		},
		{
			name: "missing description",
			raw:  `{"parentPaymentId":"pp_123"}`,
			want: incomingPaymentMetadata{},
		},
		{
			name: "malformed json silently yields zero value",
			raw:  `{"description":}`,
			want: incomingPaymentMetadata{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseIncomingPaymentMetadata(json.RawMessage(tc.raw))
			assert.Equal(t, tc.want, got)
		})
	}
}
