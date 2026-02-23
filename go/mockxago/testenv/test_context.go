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

	lastSubAccount      createSubAccountResponse
	lastBalanceResponse balanceResponse
	lastCurrencies      []currencyResponse
	previousCurrencies  string

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

	tc.lastSubAccount = createSubAccountResponse{}
	tc.lastBalanceResponse = balanceResponse{}
	tc.lastCurrencies = nil
	tc.previousCurrencies = ""

	tc.subAccountsByWallet = make(map[string]createSubAccountResponse)

	tc.lastBeneficiary = addBeneficiaryResponse{}
	tc.lastBeneficiaries = listBeneficiariesResponse{}
	tc.addedBeneficiaries = nil
}
