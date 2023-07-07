package paymentpointers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/paymentpointers"
)

func TestParsePaymentPointer(t *testing.T) {
	cases := []struct {
		name                string
		value               string
		expectedString      string
		expectedShortString string
		err                 bool
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
		{
			name:  "not_a_url",
			value: "fluffy",
			err:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pp, err := paymentpointers.Parse(tc.value)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedString, pp.String())
			assert.Equal(t, tc.expectedShortString, pp.ShortString())
		})
	}
}
