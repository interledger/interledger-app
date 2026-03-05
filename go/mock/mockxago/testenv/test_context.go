//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
	"sync"
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

	lastSubAccount       createSubAccountResponse
	lastBalanceResponse  balanceResponse
	lastCurrencies       []currencyResponse
	lastCurrenciesNested []currencyNested
	previousCurrencies   string

	subAccountsByWallet map[string]createSubAccountResponse

	// Beneficiary state
	lastBeneficiary    addBeneficiaryResponse
	lastBeneficiaries  listBeneficiariesResponse
	addedBeneficiaries []addBeneficiaryResponse

	// Transaction state
	lastTransactionID   string
	createdTransactions []string

	// Deposit/Webhook state
	webhookURL           string
	webhookServer        *http.Server
	webhookMu            sync.Mutex
	webhookEvents        []webhookEvent
	createdDeposits      []string
	lastDepositResponse  depositResponse
	lastDepositReference string
	depositRefsByWallet  map[string]map[string]string
	accountIDsByWallet   map[string]string
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

	tc.lastSubAccount = createSubAccountResponse{}
	tc.lastBalanceResponse = balanceResponse{}
	tc.lastCurrencies = nil
	tc.previousCurrencies = ""

	tc.subAccountsByWallet = make(map[string]createSubAccountResponse)

	tc.lastBeneficiary = addBeneficiaryResponse{}
	tc.lastBeneficiaries = listBeneficiariesResponse{}
	tc.addedBeneficiaries = nil

	tc.lastTransactionID = ""
	tc.createdTransactions = nil

	tc.webhookURL = ""
	tc.webhookMu.Lock()
	tc.webhookEvents = nil
	tc.webhookMu.Unlock()
	tc.createdDeposits = nil
	tc.lastDepositResponse = depositResponse{}
	tc.lastDepositReference = ""
	tc.depositRefsByWallet = make(map[string]map[string]string)
	tc.accountIDsByWallet = make(map[string]string)

	// Reset global webhook events between scenarios
	resetWebhookEvents()
}

// resetBackend calls the mockxago /v1/test/reset endpoint to clear all data
func (tc *TestContext) resetBackend() error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", tc.baseURL+"/v1/test/reset", nil)
	if err != nil {
		return err
	}
	// No auth needed for this endpoint in test mode

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
