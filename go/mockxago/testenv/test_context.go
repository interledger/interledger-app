package main

import (
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

	lastSubAccount      createSubAccountResponse
	lastBalanceResponse balanceResponse
	lastCurrencies      []currencyResponse
	previousCurrencies  string

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
}
