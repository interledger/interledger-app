package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// ============ AUTH STEPS ============

func (tc *TestContext) cleanMockGatehub() error {
	tc.Reset()
	// Reset fees to 0% for each scenario to ensure test isolation
	body := map[string]interface{}{
		"deposit_fee_percentage":    0.0,
		"withdrawal_fee_percentage": 0.0,
	}
	savedSecret := tc.appSecret
	tc.appSecret = ""
	_, err := tc.requestRaw("PUT", "/admin/fees", mustJSON(body), "application/json", nil)
	tc.appSecret = savedSecret
	if err != nil {
		return err
	}
	return nil
}

func (tc *TestContext) hmacHeaders(appID, appSecret string) error {
	tc.appID = appID
	tc.appSecret = appSecret
	return nil
}

func (tc *TestContext) postWithEmail(path, email string) error {
	tc.email = email
	body := map[string]string{"email": email}
	resp, err := tc.request("POST", path, body, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if id, ok := result["id"].(string); ok {
		tc.userID = id
	} else if id, ok := result["user_id"].(string); ok {
		tc.userID = id
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) payloadHasUserID() error {
	if tc.userID == "" {
		return fmt.Errorf("no user ID in response")
	}
	return nil
}

func (tc *TestContext) managedFlagTrue() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if managed, ok := result["managed"].(bool); !ok || !managed {
		return fmt.Errorf("expected managed to be true, got %v", result["managed"])
	}
	return nil
}

func (tc *TestContext) existingManagedUserGeneric() error {
	if tc.appSecret == "" {
		tc.hmacHeaders("local-test-app-id", "local-test-app-secret")
	}
	body := map[string]string{"email": fmt.Sprintf("user%d@example.com", time.Now().UnixNano())}
	resp, err := tc.request("POST", "/auth/v1/users/managed", body, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if id, ok := result["id"].(string); ok {
		tc.userID = id
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) existingManagedUserWithKYC(kycState string) error {
	// Create a managed user
	if err := tc.existingManagedUserGeneric(); err != nil {
		return err
	}

	// Set KYC state if needed
	if kycState == "accepted" {
		// Start KYC
		path := fmt.Sprintf("/id/v1/users/%s/hubs/gw", tc.userID)
		_, _ = tc.request("POST", path, nil, nil)

		// Submit KYC with user_id included
		submitBody := map[string]string{
			"user_id":    tc.userID,
			"first_name": "John",
			"last_name":  "Doe",
			"dob":        "1990-01-01",
			"address":    "123 Main St",
			"city":       "Anytown",
			"country":    "US",
			"risk_level": "low",
		}
		_, _ = tc.requestForm("POST", "/iframe/submit", submitBody)
	}

	return nil
}

func (tc *TestContext) authenticatedRequest() error {
	tc.hmacHeaders("local-test-app-id", "local-test-app-secret")
	return nil
}

func (tc *TestContext) managedUserFromEndpoint(path string) error {
	body := map[string]string{"email": fmt.Sprintf("user%d@example.com", time.Now().UnixNano())}
	resp, err := tc.request("POST", path, body, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if id, ok := result["id"].(string); ok {
		tc.userID = id
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) iframeTokenWithScope(scope string) error {
	scopes := parseScope(scope)
	body := map[string]interface{}{
		"scope": scopes,
	}
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	resp, err := tc.request("POST", "/auth/v1/tokens", body, headers)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if token, ok := result["token"].(string); ok {
		tc.iframeToken = token
	}

	tc.lastResponse = resp
	return nil
}
