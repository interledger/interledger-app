package paymentpointers_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/paymentpointers"
	"testing"
)

func TestParsePaymentPointer(t *testing.T) {
	cases := []struct {
		name                string
		value               string
		expectedString      string
		expectedShortString string
	}{
		{
			name:                "http",
			value:               "http://fynbos.me/asdf",
			expectedString:      "https://fynbos.me/asdf",
			expectedShortString: "fynbos.me/asdf",
		},
		{
			name:                "https",
			value:               "https://fynbos.me/asdf",
			expectedString:      "https://fynbos.me/asdf",
			expectedShortString: "fynbos.me/asdf",
		},
		{
			name:                "dollar",
			value:               "$fynbos.me/asdf",
			expectedString:      "https://fynbos.me/asdf",
			expectedShortString: "fynbos.me/asdf",
		},
		{
			name:                "noprefix",
			value:               "fynbos.me/asdf",
			expectedString:      "https://fynbos.me/asdf",
			expectedShortString: "fynbos.me/asdf",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pp, err := paymentpointers.Parse(tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedString, pp.String())
			assert.Equal(t, tc.expectedShortString, pp.ShortString())
		})
	}
}
