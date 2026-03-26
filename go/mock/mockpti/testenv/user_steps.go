//go:build e2e
// +build e2e

package main

import (
	"fmt"
)

var testUserPayload = map[string]interface{}{
	"type":        "INDIVIDUAL",
	"dateOfBirth": "1990-01-15",
	"name": map[string]string{
		"firstName": "Test",
		"lastName":  "User",
	},
	"emails": []map[string]interface{}{
		{"address": "test@example.com", "default": true},
	},
	"addresses": []map[string]interface{}{
		{
			"streetAddress": "123 Main St",
			"city":          "Springfield",
			"postalCode":    "12345",
			"stateCode":     "IL",
			"country":       "US",
			"default":       true,
		},
	},
}

func buildAssessmentPayload(userID string) map[string]interface{} {
	p := make(map[string]interface{}, len(testUserPayload)+1)
	for k, v := range testUserPayload {
		p[k] = v
	}
	p["id"] = userID
	return p
}

// postWithValidPTIUserPayload sends POST /users with a minimal valid payload.
func (tc *TestContext) postWithValidPTIUserPayload(path string) error {
	_, err := tc.ptiRequest("POST", path, testUserPayload, true)
	return err
}

// responseShouldIncludePTIUserID verifies the response has an "id" field and
// stores it for subsequent steps.
func (tc *TestContext) responseShouldIncludePTIUserID() error {
	var resp idResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if resp.ID == "" {
		return fmt.Errorf("response does not include a PTI user id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastUserID = resp.ID
	return nil
}

// anExistingPTIUser creates a user via the API and stores the id.
func (tc *TestContext) anExistingPTIUser() error {
	if err := tc.postWithValidPTIUserPayload("/users"); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludePTIUserID()
}

// postAssessmentWithScenarioID sends POST /users/assessments with an
// x-pti-scenario-id header so the mock can apply scenario-specific behaviour.
func (tc *TestContext) postAssessmentWithScenarioID(path, scenarioID string) error {
	payload := buildAssessmentPayload(tc.lastUserID)
	extraHeaders := map[string]string{
		"x-pti-scenario-id": scenarioID,
	}
	_, err := tc.ptiRequest("POST", path, payload, true, extraHeaders)
	return err
}

// responseShouldIncludeAssessmentRequestID verifies the response has a
// "requestId" field and stores it.
func (tc *TestContext) responseShouldIncludeAssessmentRequestID() error {
	var resp struct {
		ID        string `json:"id"`
		RequestID string `json:"requestId"`
	}
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	requestID := resp.RequestID
	if requestID == "" {
		requestID = resp.ID
	}
	if requestID == "" {
		return fmt.Errorf("response does not include an assessment request id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastAssessmentRequestID = requestID
	return nil
}

// responseShouldIncludeAssessment verifies the response has an "assessment" field.
func (tc *TestContext) responseShouldIncludeAssessment() error {
	return tc.responseBodyShouldIncludeField("assessment")
}

// sendGETDynamic resolves path placeholders and sends an authenticated GET.
func (tc *TestContext) sendGETDynamic(path string) error {
	if _, err := tc.ptiRequest("GET", path, nil, true); err != nil {
		return fmt.Errorf("GET %s failed: %w", path, err)
	}
	return nil
}
