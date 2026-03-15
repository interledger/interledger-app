package totp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSecretFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		want      string
		wantError bool
	}{
		{
			name: "valid otpauth URL",
			url:  "otpauth://totp/interledger.app:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=interledger.app",
			want: "JBSWY3DPEHPK3PXP",
		},
		{
			name: "valid otpauth URL with algorithm",
			url:  "otpauth://totp/interledger.app:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=interledger.app&algorithm=SHA1&digits=6&period=30",
			want: "JBSWY3DPEHPK3PXP",
		},
		{
			name:      "missing secret parameter",
			url:       "otpauth://totp/interledger.app:user@example.com?issuer=interledger.app",
			wantError: true,
		},
		{
			name:      "empty URL",
			url:       "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSecretFromURL(tt.url)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGenerateTOTPCode(t *testing.T) {
	// Test with a known secret
	secret := "JBSWY3DPEHPK3PXP" // This is "Hello!" in base32

	code, err := generateTOTPCode(secret)
	assert.NoError(t, err)
	assert.Len(t, code, 6, "TOTP code should be 6 digits")
	assert.Regexp(t, `^\d{6}$`, code, "TOTP code should be numeric")
}

func TestGenerateTOTPCode_InvalidSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"invalid base32", "!!!INVALID!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generateTOTPCode(tt.secret)
			assert.Error(t, err)
		})
	}
}

func TestExtractTOTPSecret_NoCredentials(t *testing.T) {
	client := NewClient()
	identity := &KratosIdentity{
		ID:          "test-id",
		Credentials: nil,
	}

	_, err := client.extractTOTPSecret(identity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no TOTP configured")
}

func TestExtractTOTPSecret_NoTOTPCredential(t *testing.T) {
	client := NewClient()
	identity := &KratosIdentity{
		ID: "test-id",
		Credentials: map[string]interface{}{
			"password": map[string]interface{}{},
		},
	}

	_, err := client.extractTOTPSecret(identity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no TOTP configured")
}

func TestStoreSecret(t *testing.T) {
	client := NewClient()

	// Store a valid secret
	err := client.storeSecret("test-id", "test@example.com", "JBSWY3DPEHPK3PXP")
	assert.NoError(t, err)

	// Retrieve it
	secret, found := client.getStoredSecret("test-id")
	assert.True(t, found)
	assert.Equal(t, "JBSWY3DPEHPK3PXP", secret)
}

func TestGetStoredSecret_NotFound(t *testing.T) {
	client := NewClient()

	secret, found := client.getStoredSecret("nonexistent-id")
	assert.False(t, found)
	assert.Empty(t, secret)
}
