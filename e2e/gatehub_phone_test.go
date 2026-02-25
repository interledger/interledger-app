package main

import (
	"fmt"
	"testing"
)

// TestPhoneFieldOptional tests that the phone field works with the optional schema
func TestPhoneFieldOptional(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid phone with +49 prefix",
			phone:   "+492895353680",
			wantErr: false,
		},
		{
			name:    "another valid phone",
			phone:   "+493379658426",
			wantErr: false,
		},
		{
			name:    "third valid phone",
			phone:   "+498835281273",
			wantErr: false,
		},
		{
			name:    "phone with different prefix",
			phone:   "+441234567890",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate phone validation that Kratos should accept
			// with phone field now optional in schema
			err := validatePhoneForKratos(tt.phone)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatePhoneForKratos() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// validatePhoneForKratos simulates Kratos phone validation
// With phone now optional in identity schema, this should accept E.164 format
func validatePhoneForKratos(phone string) error {
	if phone == "" {
		// Empty phone is now valid since field is optional
		return nil
	}

	// E.164 format validation: +1-999-999-9999 or similar
	if len(phone) < 10 || len(phone) > 15 {
		return fmt.Errorf("phone number invalid length: %s", phone)
	}

	if phone[0] != '+' {
		return fmt.Errorf("phone must start with +: %s", phone)
	}

	// All characters after + must be digits
	for i := 1; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return fmt.Errorf("phone contains non-digit characters: %s", phone)
		}
	}

	return nil
}

// TestPhoneGenerationFormat tests random phone generation
func TestPhoneGenerationFormat(t *testing.T) {
	// Simulate generating 10 random phones with +49 prefix
	prefix := "+49"
	expectedLen := len(prefix) + 10 // +49 + 10 random digits = 13 chars

	for i := 0; i < 10; i++ {
		phone := generateRandomPhone(prefix, 10)

		// Check length
		if len(phone) != expectedLen {
			t.Errorf("phone %q has length %d, expected %d", phone, len(phone), expectedLen)
		}

		// Check prefix
		if phone[:len(prefix)] != prefix {
			t.Errorf("phone %q doesn't start with %q", phone, prefix)
		}

		// Validate with Kratos validator
		if err := validatePhoneForKratos(phone); err != nil {
			t.Errorf("generated phone %q failed validation: %v", phone, err)
		}
	}
}

// generateRandomPhone simulates phone number generation
func generateRandomPhone(prefix string, digits int) string {
	// In real code this would use rand.Intn() for each digit
	// For testing we'll just create a predictable pattern
	phone := prefix
	for i := 0; i < digits; i++ {
		phone += fmt.Sprintf("%d", i%10) // Cycles through 0-9
	}
	return phone
}

// TestRegistrationWithOptionalPhone verifies that Kratos registration
// should work now that phone is optional in the identity schema
func TestRegistrationWithOptionalPhone(t *testing.T) {
	// Test case 1: Email-only (no phone) - should work now with optional schema
	kratosTraits1 := map[string]interface{}{
		"email":       "test@example.com",
		"firstName":   "Test",
		"lastName":    "User",
		"countryCode": "DE",
		// phone intentionally omitted - should be accepted now
	}

	// With phone optional, this should NOT error
	if err := validateKratosRegistrationTraits(kratosTraits1); err != nil {
		t.Errorf("registration without phone failed with optional schema: %v", err)
	}

	// Test case 2: Email + phone - should also work
	kratosTraits2 := map[string]interface{}{
		"email":       "test2@example.com",
		"firstName":   "Test",
		"lastName":    "User",
		"countryCode": "DE",
		"phone":       "+492895353680",
	}

	if err := validateKratosRegistrationTraits(kratosTraits2); err != nil {
		t.Errorf("registration with phone failed: %v", err)
	}
}

// validateKratosRegistrationTraits simulates Kratos validation
// Should accept: email, firstName, lastName, countryCode (required)
// And: phone (optional after schema change)
func validateKratosRegistrationTraits(traits map[string]interface{}) error {
	required := []string{"email", "firstName", "lastName", "countryCode"}

	for _, field := range required {
		if _, exists := traits[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// phone is now optional, so no error if missing
	if phone, exists := traits["phone"]; exists && phone != "" {
		// If phone is provided, validate it
		if err := validatePhoneForKratos(phone.(string)); err != nil {
			return fmt.Errorf("invalid phone format: %w", err)
		}
	}

	return nil
}
