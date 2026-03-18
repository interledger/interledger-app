package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// generateUUID generates a random UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// mockxagoLogin authenticates with MockXago and returns a bearer token.
func mockxagoLogin() (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	loginBody, err := json.Marshal(map[string]interface{}{
		"policyId": "test-policy",
		"fields": []map[string]string{
			{"fieldName": "apiPublicKey", "fieldValue": "test-public-key"},
			{"fieldName": "apiSecretKey", "fieldValue": "test-secret"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("mockxagoLogin: marshal failed: %w", err)
	}

	req, err := http.NewRequest("POST", "https://mockxago.interledger.test/v1/login", bytes.NewReader(loginBody))
	if err != nil {
		return "", fmt.Errorf("mockxagoLogin: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mockxagoLogin: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mockxagoLogin: status %d: %s", resp.StatusCode, string(body))
	}

	var loginData struct {
		TokenValue string `json:"tokenValue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginData); err != nil {
		return "", fmt.Errorf("mockxagoLogin: decode failed: %w", err)
	}
	if loginData.TokenValue == "" {
		return "", fmt.Errorf("mockxagoLogin: empty token in response")
	}

	return loginData.TokenValue, nil
}

// mockxagoClient returns an HTTP client that accepts self-signed TLS certificates.
func mockxagoClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
}

// iGetTheXagoSubAccountDetailsForTheCurrentUser retrieves sub account info from backend.
// It polls for the sub-account and linked account to exist, since the Temporal
// CreateBalanceAccountWorkflow may still be running after KYC completion.
func (sc *E2EContext) iGetTheXagoSubAccountDetailsForTheCurrentUser() error {
	debugPrintln("\n🔍 Getting Xago sub account details from backend...")

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get current user email: %w", err)
	}

	walletID, err := sc.getWalletIDByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get wallet ID: %w", err)
	}

	debugPrintf("   📍 Email: %s, Wallet ID: %s\n", email, walletID)
	sc.userDetails[sc.currentUser].Fields["xago_wallet_id"] = walletID

	// Poll for the sub-account to appear (Temporal workflow may still be running)
	var accountID string
	maxWait := 30 * time.Second
	pollInterval := 2 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		accountID, err = sc.getXagoAccountIDByWalletID(walletID)
		if err == nil && accountID != "" {
			break
		}
		debugPrintf("   ⏳ Waiting for Xago sub-account (wallet: %s)...\n", walletID)
		time.Sleep(pollInterval)
	}
	if accountID == "" {
		return fmt.Errorf("xago sub-account not created within %v for wallet %s", maxWait, walletID)
	}

	sc.userDetails[sc.currentUser].Fields["xago_account_id"] = accountID
	debugPrintf("   ✓ Found xago account ID: %s\n", accountID)

	// Also wait for the linked account to exist — the deposit webhook handler
	// needs it (ListByWalletId) to process the deposit successfully.
	linkedAccountExists := false
	for time.Now().Before(deadline) {
		exists, checkErr := sc.xagoLinkedAccountExists(walletID)
		if checkErr == nil && exists {
			debugPrintln("   ✓ Xago linked account exists")
			linkedAccountExists = true
			break
		}
		debugPrintf("   ⏳ Waiting for Xago linked account (wallet: %s)...\n", walletID)
		time.Sleep(pollInterval)
	}
	if !linkedAccountExists {
		return fmt.Errorf("xago linked account not created within %v for wallet %s", maxWait, walletID)
	}

	debugPrintln("   ✓ Sub account details retrieved")
	return nil
}

// iCreateATestTransactionInMockXagoFor creates a transaction in MockXago.
// Requires iGetTheXagoSubAccountDetailsForTheCurrentUser to have run first.
func (sc *E2EContext) iCreateATestTransactionInMockXagoFor(amountStr, currency string) error {
	debugPrintf("\n🔧 Creating test transaction in MockXago: %s %s...\n", amountStr, currency)

	walletID, exists := sc.userDetails[sc.currentUser].Fields["xago_wallet_id"]
	if !exists || walletID == "" {
		return fmt.Errorf("wallet ID not found in user details - did you call 'get the Xago sub account details' first?")
	}

	accountID, exists2 := sc.userDetails[sc.currentUser].Fields["xago_account_id"]
	if !exists2 || accountID == "" {
		return fmt.Errorf("xago_account_id not found — call 'get the Xago sub account details' first")
	}

	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	transactionID := generateUUID()

	token, err := mockxagoLogin()
	if err != nil {
		return fmt.Errorf("failed to login to MockXago: %w", err)
	}

	txReqBody, err := json.Marshal(map[string]interface{}{
		"transactionId": transactionID,
		"accountId":     accountID,
		"amount":        amount,
		"currencyCode":  currency,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal transaction request: %w", err)
	}

	client := mockxagoClient()
	req, err := http.NewRequest("POST", "https://mockxago.interledger.test/v1/test/transactions", bytes.NewReader(txReqBody))
	if err != nil {
		return fmt.Errorf("failed to create transaction request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send transaction request to MockXago: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MockXago transaction creation failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	sc.userDetails[sc.currentUser].Fields["xago_transaction_id"] = transactionID
	sc.userDetails[sc.currentUser].Fields["xago_account_id"] = accountID

	debugPrintf("   ✓ Test transaction created: %s\n", transactionID)
	return nil
}

// iPerformATestDepositOfInMockXago performs a test deposit via MockXago's test API.
func (sc *E2EContext) iPerformATestDepositOfInMockXago(amountStr, currency string) error {
	debugPrintf("\n💰 Performing test deposit: %s %s via MockXago...\n", amountStr, currency)

	walletID, exists := sc.userDetails[sc.currentUser].Fields["xago_wallet_id"]
	if !exists || walletID == "" {
		return fmt.Errorf("wallet ID not found in user details - did you call 'get the Xago sub account details' first?")
	}

	transactionID, _ := sc.userDetails[sc.currentUser].Fields["xago_transaction_id"]

	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	token, err := mockxagoLogin()
	if err != nil {
		return fmt.Errorf("failed to login to MockXago: %w", err)
	}

	depositPayload := map[string]interface{}{
		"walletId":     walletID,
		"amount":       amount,
		"currencyCode": currency,
	}

	accountID, hasAccountID := sc.userDetails[sc.currentUser].Fields["xago_account_id"]
	if hasAccountID && accountID != "" {
		depositPayload["accountId"] = accountID
		debugPrintf("   📋 Using account ID: %s\n", accountID)
	}

	if transactionID != "" {
		depositPayload["transactionId"] = transactionID
		debugPrintf("   📋 Using transaction ID: %s\n", transactionID)
	}

	depositReqBody, err := json.Marshal(depositPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal deposit request: %w", err)
	}

	debugPrintf("   📤 Sending deposit request: %s\n", string(depositReqBody))

	client := mockxagoClient()
	req, err := http.NewRequest("POST", "https://mockxago.interledger.test/v1/test/balances/deposit", bytes.NewReader(depositReqBody))
	if err != nil {
		return fmt.Errorf("failed to create deposit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send deposit request to MockXago: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	debugPrintf("   📥 MockXago response: status=%d, body=%s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MockXago deposit failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	debugPrintf("   ✓ Deposit accepted by MockXago\n")
	return nil
}

// iWaitSecondsForTheWebhookToBeProcessed waits for the webhook to be delivered and processed.
func (sc *E2EContext) iWaitSecondsForTheWebhookToBeProcessed(secondsStr string) error {
	var seconds int
	if _, err := fmt.Sscanf(secondsStr, "%d", &seconds); err != nil {
		return fmt.Errorf("invalid wait time: %s", secondsStr)
	}

	debugPrintf("\n⏳ Waiting %d seconds for webhook processing...\n", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)

	debugPrintln("   ✓ Wait complete, webhook should be processed")
	return nil
}

// iInitiateADepositForMyXagoLinkedAccount navigates to the deposit page,
// retrieves sub account details, and takes a screenshot.
func (sc *E2EContext) iInitiateADepositForMyXagoLinkedAccount() error {
	debugPrintln("\n🏦 Initiating deposit for Xago linked account...")

	if err := sc.iGetTheXagoSubAccountDetailsForTheCurrentUser(); err != nil {
		return fmt.Errorf("failed to get Xago sub account details: %w", err)
	}

	if err := sc.verifyMockXagoBankAccountExists(); err != nil {
		return fmt.Errorf("MockXago bank account not configured: %w", err)
	}

	if err := sc.iNavigateToTheDepositPage(); err != nil {
		return fmt.Errorf("failed to navigate to deposit page: %w", err)
	}

	debugPrintln("   ⏳ Waiting for deposit instructions to load...")
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(2 * time.Second)

	_ = sc.iTakeAScreenshot("xago_deposit_instructions")
	debugPrintln("   ✓ Deposit initiated, instructions should be visible")
	return nil
}

// verifyMockXagoBankAccountExists checks that MockXago has bank accounts configured for ZAR.
func (sc *E2EContext) verifyMockXagoBankAccountExists() error {
	token, err := mockxagoLogin()
	if err != nil {
		return fmt.Errorf("failed to login to MockXago: %w", err)
	}

	client := mockxagoClient()
	req, err := http.NewRequest("GET", "https://mockxago.interledger.test/v1/currencies", nil)
	if err != nil {
		return fmt.Errorf("failed to create currencies request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get banking providers: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("currencies endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var bankAccounts []map[string]interface{}
	if err := json.Unmarshal(body, &bankAccounts); err != nil {
		return fmt.Errorf("failed to parse banking providers response: %w", err)
	}

	for _, account := range bankAccounts {
		if currency, ok := account["currencyCode"].(string); ok && currency == "ZAR" {
			if depositEnabled, ok := account["depositEnabled"].(bool); ok && depositEnabled {
				debugPrintln("   ✓ Found ZAR bank account with deposits enabled")
				return nil
			}
		}
	}

	return fmt.Errorf("no ZAR bank account with deposits enabled found in MockXago")
}

// myXagoSpecificDepositInstructionsShouldBeDisplayedToMe verifies that
// the Xago deposit instructions are displayed with bank details.
func (sc *E2EContext) myXagoSpecificDepositInstructionsShouldBeDisplayedToMe() error {
	debugPrintln("\n🔍 Verifying Xago deposit instructions are displayed...")

	time.Sleep(2 * time.Second)

	pageContent, err := sc.page.TextContent("body")
	if err != nil {
		return fmt.Errorf("failed to get page content: %w", err)
	}

	errorPatterns := []string{
		"500",
		"Internal Server Error",
		"Internal server error",
		"Something went wrong",
		"Error loading",
		"Failed to load",
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(pageContent, pattern) {
			_ = sc.iTakeAScreenshot("xago_deposit_page_error")
			return fmt.Errorf("page shows error (%s) instead of deposit instructions", pattern)
		}
	}

	fieldSelectors := []struct {
		name     string
		selector string
	}{
		{"Bank field", "text=Bank"},
		{"Branch code field", "text=Branch code"},
		{"Account number field", "text=Account number"},
		{"Reference field", "text=Reference"},
	}

	maxWait := 10 * time.Second
	deadline := time.Now().Add(maxWait)

	for _, field := range fieldSelectors {
		found := false
		for time.Now().Before(deadline) {
			locator := sc.page.Locator(field.selector)
			count, _ := locator.Count()
			if count > 0 {
				debugPrintf("   ✓ %s is displayed\n", field.name)
				found = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if !found {
			_ = sc.iTakeAScreenshot("xago_deposit_instructions_missing_fields")
			return fmt.Errorf("%s not found on page - deposit instructions may not have loaded", field.name)
		}
	}

	debugPrintln("   ✓ All Xago deposit instructions are displayed correctly")
	return nil
}

// iSimulateAnEFTPaymentToXago simulates an EFT payment by creating a transaction
// in MockXago and performing a test deposit.
func (sc *E2EContext) iSimulateAnEFTPaymentToXago(amountStr, currency string) error {
	debugPrintf("\n💳 Simulating EFT payment of %s %s to Xago...\n", amountStr, currency)

	if err := sc.iCreateATestTransactionInMockXagoFor(amountStr, currency); err != nil {
		return fmt.Errorf("failed to create test transaction: %w", err)
	}

	if err := sc.iPerformATestDepositOfInMockXago(amountStr, currency); err != nil {
		return fmt.Errorf("failed to perform test deposit: %w", err)
	}

	debugPrintf("   ✓ EFT payment simulation complete for %s %s\n", amountStr, currency)
	return nil
}
