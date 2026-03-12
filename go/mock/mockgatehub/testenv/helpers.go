//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============ HELPER FUNCTIONS ============

// parseScope parses a scope string (can be comma-separated or single value)
func parseScope(scope string) []string {
	scopes := strings.Split(strings.Trim(scope, "\""), ",")
	for i := range scopes {
		scopes[i] = strings.TrimSpace(scopes[i])
		scopes[i] = strings.Trim(scopes[i], "\"")
	}
	return scopes
}

// checkRequiredFields validates that the response contains all required fields
func (tc *TestContext) checkRequiredFields(fields []string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	for _, field := range fields {
		if _, ok := result[field]; !ok {
			return fmt.Errorf("missing field %s", field)
		}
	}

	return nil
}

// checkFieldValue validates that a specific field has the expected value
func (tc *TestContext) checkFieldValue(field, expectedValue string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	value, ok := result[field]
	if !ok {
		return fmt.Errorf("missing field %s", field)
	}

	valueStr := fmt.Sprintf("%v", value)
	if valueStr != expectedValue {
		return fmt.Errorf("expected %s to be %s, got %v", field, expectedValue, value)
	}

	return nil
}

// ============ BASIC STEPS (Common to all features) ============

func (tc *TestContext) mockgatehubStarted() error {
	tc.Reset()
	return nil
}

func (tc *TestContext) serviceURLIs(url string) error {
	tc.baseURL = url
	return nil
}

func (tc *TestContext) sendGetRequestWithHMAC(path string) error {
	tc.appID = "local-test-app-id"
	tc.appSecret = "local-test-app-secret"
	resp, err := tc.request("GET", path, nil, nil)
	if err != nil {
		tc.lastError = err
		return err
	}
	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) responseStatusIs(status int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response")
	}
	if tc.lastResponse.StatusCode != status {
		return fmt.Errorf("expected status %d, got %d. Body: %s", status, tc.lastResponse.StatusCode, string(tc.lastResponseBody))
	}
	return nil
}

func (tc *TestContext) responseBodyContains(key, value string) error {
	var data map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &data); err != nil {
		return err
	}
	if v, ok := data[key]; !ok || fmt.Sprintf("%v", v) != value {
		return fmt.Errorf("expected %s to be %s, got %v", key, value, v)
	}
	return nil
}
