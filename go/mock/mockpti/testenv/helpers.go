//go:build e2e
// +build e2e

package main

import (
	"encoding/json"
	"fmt"
)

func (tc *TestContext) decodeLastResponse(out interface{}) error {
	if tc.lastResponseBody == nil {
		return fmt.Errorf("no response body")
	}
	return json.Unmarshal(tc.lastResponseBody, out)
}

func (tc *TestContext) responseStatusShouldBe(status int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response")
	}
	if tc.lastResponse.StatusCode != status {
		return fmt.Errorf("expected status %d, got %d. Body: %s", status, tc.lastResponse.StatusCode, string(tc.lastResponseBody))
	}
	return nil
}

func (tc *TestContext) responseBodyShouldContainAs(field, value string) error {
	var body map[string]interface{}
	if err := tc.decodeLastResponse(&body); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	got, ok := body[field]
	if !ok {
		return fmt.Errorf("response body does not contain field %q", field)
	}
	if fmt.Sprintf("%v", got) != value {
		return fmt.Errorf("expected field %q to be %q, got %q", field, value, got)
	}
	return nil
}

func (tc *TestContext) responseBodyShouldIncludeField(field string) error {
	var body map[string]interface{}
	if err := tc.decodeLastResponse(&body); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	if _, ok := body[field]; !ok {
		return fmt.Errorf("response body does not include field %q. Body: %s", field, string(tc.lastResponseBody))
	}
	return nil
}

func (tc *TestContext) responseBodyShouldIncludeEqual(field, placeholder string) error {
	expected := tc.resolvePlaceholder(placeholder)

	var body map[string]interface{}
	if err := tc.decodeLastResponse(&body); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	got, ok := body[field]
	if !ok {
		return fmt.Errorf("response body does not include field %q", field)
	}
	if fmt.Sprintf("%v", got) != expected {
		return fmt.Errorf("expected field %q to be %q, got %q", field, expected, got)
	}
	return nil
}

// resolvePlaceholder replaces step-level placeholders like {userId} with stored values.
func (tc *TestContext) resolvePlaceholder(s string) string {
	switch s {
	case "{userId}":
		return tc.lastUserID
	case "{walletId}":
		return tc.lastWalletID
	case "{paymentInformationId}":
		return tc.lastPaymentInformationID
	}
	return s
}
