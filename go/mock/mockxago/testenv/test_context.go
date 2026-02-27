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

	// Sub-account state
	lastSubAccount      createSubAccountResponse
	subAccountsByWallet map[string]createSubAccountResponse

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

	tc.lastSubAccount = createSubAccountResponse{}
	tc.subAccountsByWallet = make(map[string]createSubAccountResponse)

	tc.lastBeneficiary = addBeneficiaryResponse{}
	tc.lastBeneficiaries = listBeneficiariesResponse{}
	tc.addedBeneficiaries = nil

	// Reset global webhook events between scenarios
	// resetWebhookEvents()
}

// resetBackend calls the mockxago /v1/test/reset endpoint to clear all data
func (tc *TestContext) resetBackend() error {
	req, err := http.NewRequest("POST", tc.baseURL+"/v1/test/reset", nil)
	if err != nil {
		return err
	}
	resp, err := tc.client.Do(req)
	if err != nil {
		// Server might not be ready yet; reset client-side only
		tc.Reset()
		return nil
	}
	resp.Body.Close()
	tc.Reset()
	return nil
}
