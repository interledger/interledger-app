package main

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// ============ ORGANIZATION CONFIGURATION STEPS ============

func (tc *TestContext) patchOrganizationConfig(orgID, apiBaseURL, type2fa string) error {
	body := map[string]string{
		"apiBaseUrl": apiBaseURL,
		"type2fa":    type2fa,
	}

	resp, err := tc.request("PATCH", "/auth/v1/users/organization/"+orgID, body, nil)
	if err != nil {
		return err
	}
	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) responseFieldIs(field, expected string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	actual, ok := result[field]
	if !ok {
		return fmt.Errorf("field %q not found in response. Available: %v", field, result)
	}

	actualStr := fmt.Sprintf("%v", actual)
	if actualStr != expected {
		return fmt.Errorf("expected %s to be %q, got %q", field, expected, actualStr)
	}

	return nil
}

func (tc *TestContext) responseHasTimestamps() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	for _, field := range []string{"createdAt", "updatedAt"} {
		val, ok := result[field]
		if !ok {
			return fmt.Errorf("missing field %q in response", field)
		}
		str, ok := val.(string)
		if !ok || str == "" {
			return fmt.Errorf("field %q should be a non-empty string, got %v", field, val)
		}
	}

	return nil
}

func InitializeOrganizationScenario(ctx *godog.ScenarioContext, tc *TestContext) {
	ctx.Step(`^I PATCH /auth/v1/users/organization/([^ ]+) with apiBaseUrl "([^"]*)" and type2fa "([^"]*)"$`, tc.patchOrganizationConfig)
	ctx.Step(`^the response field "([^"]*)" is "([^"]*)"$`, tc.responseFieldIs)
	ctx.Step(`^the response has "createdAt" and "updatedAt" timestamps$`, tc.responseHasTimestamps)
}
