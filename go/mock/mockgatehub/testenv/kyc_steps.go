//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============ KYC STEPS ============

func (tc *TestContext) postEndpoint(path string) error {
	resp, err := tc.request("POST", path, nil, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response from %s: %w. Response body: %s. Status: %d", path, err, string(tc.lastResponseBody), resp.StatusCode)
	}

	if token, ok := result["token"].(string); ok {
		tc.kycToken = token
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) responseHasKYCToken() error {
	if tc.kycToken == "" {
		return fmt.Errorf("no KYC token in response")
	}
	return nil
}

func (tc *TestContext) getShowsKYCState(path, kycState, riskLevel string) error {
	resp, err := tc.request("GET", path, nil, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if state, ok := result["kyc_state"].(string); !ok || state != kycState {
		return fmt.Errorf("expected kyc_state %s, got %v", kycState, result["kyc_state"])
	}

	if risk, ok := result["risk_level"].(string); !ok || risk != riskLevel {
		return fmt.Errorf("expected risk_level %s, got %v", riskLevel, result["risk_level"])
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) getEndpoint(path string) error {
	resp, err := tc.request("GET", path, nil, nil)
	if err != nil {
		return err
	}
	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) responseIsHTMLMentioning(text1, text2 string) error {
	if !strings.Contains(string(tc.lastResponseBody), text1) || !strings.Contains(string(tc.lastResponseBody), text2) {
		return fmt.Errorf("expected HTML to mention %s and %s", text1, text2)
	}
	return nil
}

func (tc *TestContext) submitKYCFormWithout2FA() error {
	formData := map[string]string{
		"user_id":     tc.userID,
		"first_name":  "Jane",
		"last_name":   "Doe",
		"dob":         "1990-05-15",
		"address":     "123 Main St",
		"city":        "Testville",
		"country":     "US",
		"risk_level":  "low",
		"kyc_outcome": "accepted",
	}
	_, err := tc.requestForm("POST", "/iframe/submit", formData)
	return err
}

func (tc *TestContext) submitKYCFormWithOutcome(outcome string) error {
	formData := map[string]string{
		"user_id":     tc.userID,
		"first_name":  "Jane",
		"last_name":   "Doe",
		"dob":         "1990-05-15",
		"address":     "123 Main St",
		"city":        "Testville",
		"country":     "US",
		"risk_level":  "low",
		"kyc_outcome": outcome,
	}
	_, err := tc.requestForm("POST", "/iframe/submit", formData)
	return err
}

func (tc *TestContext) submitKYCFormWith2FA(code string) error {
	formData := map[string]string{
		"user_id":     tc.userID,
		"first_name":  "Jane",
		"last_name":   "Doe",
		"dob":         "1990-05-15",
		"address":     "123 Main St",
		"city":        "Testville",
		"country":     "US",
		"risk_level":  "low",
		"trigger_2fa": "on",
		"totp_code":   code,
	}
	_, err := tc.requestForm("POST", "/iframe/submit", formData)
	return err
}
