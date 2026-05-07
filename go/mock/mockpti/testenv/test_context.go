//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
	"time"
)

// TestContext holds state for godog scenarios.
type TestContext struct {
	client  *http.Client
	baseURL string

	// PTI auth header value
	clientID string

	lastResponse     *http.Response
	lastResponseBody []byte
	lastError        error

	// Stored IDs for use across steps within a scenario
	lastUserID               string
	lastAssessmentRequestID  string
	lastWalletID             string
	lastPaymentInformationID string
	lastTransactionRequestID string
	depositRequestID         string
	withdrawalRequestID      string
	depositAmount            float64
	lastUpdateID             string
	lastWebhook         map[string]interface{}
	lastWebhookSigned   bool
}

// Reset initializes the test context to a clean state.
func (tc *TestContext) Reset() {
	tc.client = &http.Client{Timeout: 10 * time.Second}
	tc.baseURL = mockPTIURL
	tc.clientID = defaultClientID

	tc.lastResponse = nil
	tc.lastResponseBody = nil
	tc.lastError = nil

	tc.lastUserID = ""
	tc.lastAssessmentRequestID = ""
	tc.lastWalletID = ""
	tc.lastPaymentInformationID = ""
	tc.lastTransactionRequestID = ""
	tc.depositRequestID = ""
	tc.withdrawalRequestID = ""
	tc.depositAmount = 0
	tc.lastUpdateID = ""
	tc.lastWebhook = nil
	tc.lastWebhookSigned = false
}

// resetBackend calls the /test/reset endpoint to clear all mock data.
func (tc *TestContext) resetBackend() error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", tc.baseURL+"/test/reset", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend reset failed with status: %d", resp.StatusCode)
	}
	return nil
}
