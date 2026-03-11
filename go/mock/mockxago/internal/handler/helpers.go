package handler

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

var (
	// ErrMissingAuthHeader is returned when Authorization header is missing
	ErrMissingAuthHeader = errors.New("missing authorization header")
	// ErrInvalidAuthFormat is returned when Authorization header format is invalid
	ErrInvalidAuthFormat = errors.New("invalid authorization format")
)

// BeneficiaryToItem converts a Beneficiary model to API response item
// This function is exported to make it independently testable
func BeneficiaryToItem(b *models.Beneficiary) models.BeneficiaryItem {
	return models.BeneficiaryItem{
		UUID:          b.ID,
		Name:          b.Name,
		Scope:         b.Scope,
		CurrencyCode:  b.Currency,
		AccountNumber: b.AccountNumber,
		BranchCode:    b.BranchCode,
		BankName:      b.BankName,
		AccountName:   b.AccountName,
		Reference:     b.Reference,
		IsOwn:         b.IsOwn,
		Status:        b.Status,
	}
}

// ParsePagination extracts limit and page from query params with defaults
// Returns limit, page, and calculated offset
func ParsePagination(limitStr, pageStr string, defaultLimit int) (limit, page, offset int) {
	limit = defaultLimit
	page = 1

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	offset = (page - 1) * limit
	return
}

// CalculatePages returns number of pages for pagination
func CalculatePages(total, limit int) int {
	if total == 0 || limit == 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}

// ExtractBearerToken parses Authorization header and returns token value
// Returns error if header is missing or format is invalid
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	// Use SplitN to handle multiple spaces gracefully
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidAuthFormat
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrInvalidAuthFormat
	}

	return token, nil
}

// FormatSettledAt formats a time for transaction settled_at field
// Uses RFC3339 format in UTC timezone
func FormatSettledAt(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ValidateRequiredString checks if a string field is non-empty after trimming
func ValidateRequiredString(value, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New(fieldName + " is required")
	}
	return trimmed, nil
}
