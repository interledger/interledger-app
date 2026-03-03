//go:build e2e

package main

import (
	"net/http"
	"time"
)

// TestContext holds state for godog scenarios.
type TestContext struct {
	client *http.Client

	baseURL string
	policy  string
	pubKey  string
	secret  string
	token   string

	lastResponse     *http.Response
	lastResponseBody []byte
	lastError        error

	lastLoginToken string
	expiredToken   string

	lastWalletID      string
	lastEmail         string
	previousAccountID string

	previousCurrencies string

	// Beneficiary state
	lastBeneficiary    addBeneficiaryResponse
	lastBeneficiaries  listBeneficiariesResponse
	addedBeneficiaries []addBeneficiaryResponse
}

// Reset initializes the test context to a clean state.
func (tc *TestContext) Reset() {
	tc.client = &http.Client{Timeout: 10 * time.Second}
	tc.baseURL = mockXagoURL
	tc.policy = defaultPolicy
	tc.pubKey = defaultPubKey
	tc.secret = defaultSecret
	tc.token = ""

	tc.lastResponse = nil
	tc.lastResponseBody = nil
	tc.lastError = nil

	tc.lastLoginToken = ""
	tc.expiredToken = ""

	tc.lastWalletID = ""
	tc.lastEmail = ""
	tc.previousAccountID = ""

	tc.previousCurrencies = ""

	tc.lastBeneficiary = addBeneficiaryResponse{}
	tc.lastBeneficiaries = listBeneficiariesResponse{}
	tc.addedBeneficiaries = nil

	// Reset global webhook events between scenarios
	// resetWebhookEvents()
}

// resetBackend calls the mockxago /v1/test/reset endpoint to clear all data
func (tc *TestContext) resetBackend() error {
	// For the minimal authentication-only version, we simply reset the client state
	tc.Reset()
	return nil
}
