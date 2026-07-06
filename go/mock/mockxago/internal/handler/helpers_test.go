package handler

import (
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestBeneficiaryToItem(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.Beneficiary
		expected models.BeneficiaryItem
	}{
		{
			name: "full beneficiary with all fields",
			input: &models.Beneficiary{
				ID:            "ben-123",
				AccountID:     "acc-456",
				WalletID:      "wallet-789",
				Name:          "John Doe",
				Scope:         "domestic",
				IsOwn:         true,
				BankName:      "Test Bank",
				AccountNumber: "1234567890",
				AccountName:   "John Doe Account",
				BranchCode:    "123456",
				Reference:     "Salary",
				Currency:      "ZAR",
				Status:        "approved",
			},
			expected: models.BeneficiaryItem{
				UUID:          "ben-123",
				Name:          "John Doe",
				Scope:         "domestic",
				CurrencyCode:  "ZAR",
				AccountNumber: "1234567890",
				BranchCode:    "123456",
				BankName:      "Test Bank",
				AccountName:   "John Doe Account",
				Reference:     "Salary",
				IsOwn:         true,
				Status:        "approved",
			},
		},
		{
			name: "minimal beneficiary",
			input: &models.Beneficiary{
				ID:            "ben-min",
				Name:          "Minimal Ben",
				AccountNumber: "9876543210",
				Currency:      "USD",
				Status:        "pending",
			},
			expected: models.BeneficiaryItem{
				UUID:          "ben-min",
				Name:          "Minimal Ben",
				AccountNumber: "9876543210",
				CurrencyCode:  "USD",
				Status:        "pending",
			},
		},
		{
			name: "beneficiary with empty optional fields",
			input: &models.Beneficiary{
				ID:            "ben-empty",
				Name:          "Empty Fields",
				AccountNumber: "1111111111",
				Currency:      "EUR",
				Status:        "rejected",
				Scope:         "",
				BankName:      "",
				AccountName:   "",
				BranchCode:    "",
				Reference:     "",
			},
			expected: models.BeneficiaryItem{
				UUID:          "ben-empty",
				Name:          "Empty Fields",
				AccountNumber: "1111111111",
				CurrencyCode:  "EUR",
				Status:        "rejected",
				Scope:         "",
				BankName:      "",
				AccountName:   "",
				BranchCode:    "",
				Reference:     "",
				IsOwn:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BeneficiaryToItem(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name           string
		limitStr       string
		pageStr        string
		defaultLimit   int
		expectedLimit  int
		expectedPage   int
		expectedOffset int
	}{
		{
			name:           "valid limit and page",
			limitStr:       "20",
			pageStr:        "3",
			defaultLimit:   10,
			expectedLimit:  20,
			expectedPage:   3,
			expectedOffset: 40, // (3-1) * 20
		},
		{
			name:           "empty strings use defaults",
			limitStr:       "",
			pageStr:        "",
			defaultLimit:   25,
			expectedLimit:  25,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "invalid limit falls back to default",
			limitStr:       "invalid",
			pageStr:        "2",
			defaultLimit:   15,
			expectedLimit:  15,
			expectedPage:   2,
			expectedOffset: 15,
		},
		{
			name:           "invalid page falls back to 1",
			limitStr:       "30",
			pageStr:        "not-a-number",
			defaultLimit:   10,
			expectedLimit:  30,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "negative limit uses default",
			limitStr:       "-10",
			pageStr:        "5",
			defaultLimit:   50,
			expectedLimit:  50,
			expectedPage:   5,
			expectedOffset: 200,
		},
		{
			name:           "zero limit uses default",
			limitStr:       "0",
			pageStr:        "1",
			defaultLimit:   100,
			expectedLimit:  100,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "negative page uses default",
			limitStr:       "5",
			pageStr:        "-1",
			defaultLimit:   10,
			expectedLimit:  5,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "zero page uses default",
			limitStr:       "8",
			pageStr:        "0",
			defaultLimit:   10,
			expectedLimit:  8,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "page 1 has offset 0",
			limitStr:       "100",
			pageStr:        "1",
			defaultLimit:   10,
			expectedLimit:  100,
			expectedPage:   1,
			expectedOffset: 0,
		},
		{
			name:           "large page number",
			limitStr:       "25",
			pageStr:        "42",
			defaultLimit:   10,
			expectedLimit:  25,
			expectedPage:   42,
			expectedOffset: 1025, // (42-1) * 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, page, offset := ParsePagination(tt.limitStr, tt.pageStr, tt.defaultLimit)
			assert.Equal(t, tt.expectedLimit, limit, "limit mismatch")
			assert.Equal(t, tt.expectedPage, page, "page mismatch")
			assert.Equal(t, tt.expectedOffset, offset, "offset mismatch")
		})
	}
}

func TestCalculatePages(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		limit    int
		expected int
	}{
		{name: "exact division", total: 100, limit: 10, expected: 10},
		{name: "with remainder", total: 101, limit: 10, expected: 11},
		{name: "total less than limit", total: 5, limit: 10, expected: 1},
		{name: "zero total", total: 0, limit: 10, expected: 1},
		{name: "zero limit", total: 100, limit: 0, expected: 1},
		{name: "one item one page", total: 1, limit: 1, expected: 1},
		{name: "large total", total: 9999, limit: 25, expected: 400},
		{name: "prime numbers", total: 97, limit: 13, expected: 8},
		{name: "both zero", total: 0, limit: 0, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePages(tt.total, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectedErr   error
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer abc123xyz",
			expectedToken: "abc123xyz",
			expectedErr:   nil,
		},
		{
			name:          "valid with extra spaces",
			authHeader:    "Bearer   token-with-spaces   ",
			expectedToken: "token-with-spaces",
			expectedErr:   nil,
		},
		{
			name:          "lowercase bearer",
			authHeader:    "bearer mytoken",
			expectedToken: "mytoken",
			expectedErr:   nil,
		},
		{
			name:          "mixed case bearer",
			authHeader:    "BeArEr MixedCaseToken",
			expectedToken: "MixedCaseToken",
			expectedErr:   nil,
		},
		{
			name:          "empty header",
			authHeader:    "",
			expectedToken: "",
			expectedErr:   ErrMissingAuthHeader,
		},
		{
			name:          "missing Bearer prefix",
			authHeader:    "justtoken",
			expectedToken: "",
			expectedErr:   ErrInvalidAuthFormat,
		},
		{
			name:          "wrong prefix",
			authHeader:    "Basic abc123",
			expectedToken: "",
			expectedErr:   ErrInvalidAuthFormat,
		},
		{
			name:          "only Bearer no token",
			authHeader:    "Bearer",
			expectedToken: "",
			expectedErr:   ErrInvalidAuthFormat,
		},
		{
			name:          "Bearer with empty token",
			authHeader:    "Bearer   ",
			expectedToken: "",
			expectedErr:   ErrInvalidAuthFormat,
		},
		{
			name:          "multiple spaces between Bearer and token",
			authHeader:    "Bearer     my-token-123",
			expectedToken: "my-token-123",
			expectedErr:   nil,
		},
		{
			name:          "token with special characters",
			authHeader:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tt.authHeader)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestFormatSettledAt(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "specific UTC time",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
			expected: "2024-03-15T14:30:45Z",
		},
		{
			name:     "converts non-UTC to UTC",
			input:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("EST", -5*3600)),
			expected: "2024-06-01T17:00:00Z", // 12:00 EST is 17:00 UTC
		},
		{
			name:     "midnight",
			input:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2025-01-01T00:00:00Z",
		},
		{
			name:     "end of day",
			input:    time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2023-12-31T23:59:59Z",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSettledAt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRequiredString(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		fieldName   string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid non-empty string",
			value:       "John Doe",
			fieldName:   "name",
			expected:    "John Doe",
			expectError: false,
		},
		{
			name:        "trims whitespace",
			value:       "  trimmed value  ",
			fieldName:   "field",
			expected:    "trimmed value",
			expectError: false,
		},
		{
			name:        "empty string",
			value:       "",
			fieldName:   "email",
			expected:    "",
			expectError: true,
			errorMsg:    "email is required",
		},
		{
			name:        "only whitespace",
			value:       "     ",
			fieldName:   "address",
			expected:    "",
			expectError: true,
			errorMsg:    "address is required",
		},
		{
			name:        "tabs and newlines",
			value:       "\t\n\r",
			fieldName:   "description",
			expected:    "",
			expectError: true,
			errorMsg:    "description is required",
		},
		{
			name:        "single character",
			value:       "x",
			fieldName:   "code",
			expected:    "x",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateRequiredString(tt.value, tt.fieldName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
