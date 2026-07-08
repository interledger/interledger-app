package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgreementDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"known: privacy_policy", "privacy_policy", "Privacy Policy"},
		{"known: terms_of_service", "terms_of_service", "Terms of Service"},
		{"known: user_policy", "user_policy", "User Policy"},
		{"unknown: single word", "disclaimer", "Disclaimer"},
		{"unknown: multi-word", "cookie_notice", "Cookie Notice"},
		{"unknown: preserves title-case only", "DATA_SHARING", "Data Sharing"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, agreementDisplayName(tc.input))
		})
	}
}
