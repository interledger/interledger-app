package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var supportedPaymentCurrencies = map[string]struct{}{
	"USD": {},
	"NGN": {},
	"CAD": {},
}

func parseAmountString(raw string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("amount is required")
	}

	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount")
	}
	if v <= 0 {
		return 0, fmt.Errorf("amount must be greater than 0")
	}
	return v, nil
}

func generateIssueID(subAccount string) string {
	sub := strings.TrimSpace(subAccount)
	if sub == "" {
		sub = "mockchimoney-root"
	}
	return sub + "_" + uuid.NewString()
}

func appendQuery(rawURL string, params map[string]string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func parseFlexibleFloat(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("missing value")
	}

	var asNumber float64
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber, nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strconv.ParseFloat(strings.TrimSpace(asString), 64)
	}

	return 0, fmt.Errorf("invalid numeric value")
}
