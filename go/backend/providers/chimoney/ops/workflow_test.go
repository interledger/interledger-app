package ops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertPhoneToE164(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "phone with dashes",
			input:    "+1-604-555-0199",
			expected: "+16045550199",
		},
		{
			name:     "phone with spaces",
			input:    "+1 604 555 0199",
			expected: "+16045550199",
		},
		{
			name:     "phone already in E164 format",
			input:    "+16045550199",
			expected: "+16045550199",
		},
		{
			name:     "phone with parentheses",
			input:    "+1 (604) 555-0199",
			expected: "+16045550199",
		},
		{
			name:     "phone with dots",
			input:    "+1.604.555.0199",
			expected: "+16045550199",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertPhoneToE164(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
