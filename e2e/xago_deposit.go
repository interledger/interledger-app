package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	mockXagoBaseURL = "https://mockxago.interledger.test"
	xagoPublicKey   = "test-public-key"
	xagoSecretKey   = "test-secret"
	xagoPolicyID    = "test-policy"
)

// xagoTokenCache caches the MockXago auth token
var xagoTokenCache struct {
	token  string
	expiry time.Time
}

// xagoHTTPClient returns an HTTP client that accepts self-signed TLS certificates
func xagoHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// getMockXagoAuthToken authenticates with MockXago and returns a bearer token.
// Tokens are cached to avoid repeated login calls.
func getMockXagoAuthToken() (string, error) {
	if xagoTokenCache.token != "" && time.Now().Before(xagoTokenCache.expiry) {
		return xagoTokenCache.token, nil
	}

	debugPrintln("   🔑 Authenticating with MockXago...")

	payload := map[string]interface{}{
		"policyId": xagoPolicyID,
		"fields": []map[string]string{
			{"fieldName": "apiPublicKey", "fieldValue": xagoPublicKey},
			{"fieldName": "apiSecretKey", "fieldValue": xagoSecretKey},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("getMockXagoAuthToken: marshal failed: %w", err)
	}

	client := xagoHTTPClient()
	req, err := http.NewRequest("POST", mockXagoBaseURL+"/v1/login", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("getMockXagoAuthToken: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("getMockXagoAuthToken: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("getMockXagoAuthToken: MockXago login returned status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		TokenValue string `json:"tokenValue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("getMockXagoAuthToken: decode response failed: %w", err)
	}

	if loginResp.TokenValue == "" {
		return "", fmt.Errorf("getMockXagoAuthToken: empty token in response")
	}

	xagoTokenCache.token = loginResp.TokenValue
	xagoTokenCache.expiry = time.Now().Add(50 * time.Minute)
	debugPrintf("   ✓ Got MockXago auth token\n")
	return xagoTokenCache.token, nil
}

// getXagoAccountIDByEmail retrieves the Xago sub-account account_id for a user by email.
// It queries the xago_sub_accounts table via the user's wallet.
func (sc *E2EContext) getXagoAccountIDByEmail(email string) (string, error) {
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("getXagoAccountIDByEmail: failed to open db: %w", err)
		}
		sc.db = db
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return "", fmt.Errorf("getXagoAccountIDByEmail: could not resolve kratos user id for %s", email)
	}

	debugPrintf("   📋 Looking up Xago account ID for email: %s (kratos: %s)\n", email, kratosID)

	var accountID string
	err := sc.db.QueryRow(`
		SELECT xsa.account_id
		FROM xago_sub_accounts xsa
		JOIN user_wallets uw ON uw.wallet_id = xsa.wallet_id
		WHERE uw.user_id = $1
		LIMIT 1
	`, kratosID).Scan(&accountID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getXagoAccountIDByEmail: no Xago sub-account found for user %s", email)
		}
		return "", fmt.Errorf("getXagoAccountIDByEmail: database error: %w", err)
	}

	debugPrintf("   ✓ Found Xago account ID: %s\n", accountID)
	return accountID, nil
}

// iSimulateXagoTestDeposit calls MockXago's testdeposit endpoint to simulate a deposit.
// This triggers the full webhook flow: MockXago processes the deposit, sends a webhook
// to the backend, which updates the user's balance.
func (sc *E2EContext) iSimulateXagoTestDeposit(amount, currency string) error {
	debugPrintf("\n💰 Simulating Xago test deposit of %s %s...\n", amount, currency)

	// Get current user's email
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: failed to get current user email: %w", err)
	}

	// Poll for the Xago account ID (KYC flow may still be completing)
	var accountID string
	for attempt := 0; attempt < 20; attempt++ {
		accountID, err = sc.getXagoAccountIDByEmail(email)
		if err == nil {
			break
		}
		debugPrintf("   ⏳ Waiting for Xago sub-account to be created (attempt %d/20)...\n", attempt+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: Xago sub-account not found after waiting: %w", err)
	}

	// Get auth token for MockXago
	token, err := getMockXagoAuthToken()
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: failed to get MockXago auth token: %w", err)
	}

	// Parse amount
	var amountFloat float64
	if _, err := fmt.Sscanf(amount, "%f", &amountFloat); err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: invalid amount %q: %w", amount, err)
	}

	// Build testdeposit payload
	payload := map[string]interface{}{
		"accountId":    accountID,
		"amount":       amountFloat,
		"currencyCode": strings.ToUpper(currency),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: marshal failed: %w", err)
	}

	// POST to MockXago testdeposit endpoint
	client := xagoHTTPClient()
	req, err := http.NewRequest("POST", mockXagoBaseURL+"/v1/company/accounts/testdeposit", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("iSimulateXagoTestDeposit: MockXago testdeposit returned status %d: %s", resp.StatusCode, string(body))
	}

	var depositResp struct {
		TransactionID string `json:"transactionId"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(body, &depositResp); err != nil {
		return fmt.Errorf("iSimulateXagoTestDeposit: decode response failed: %w", err)
	}

	debugPrintf("   ✓ Xago test deposit initiated: txn=%s status=%s\n", depositResp.TransactionID, depositResp.Status)
	_ = sc.iTakeAScreenshot("xago-deposit-initiated")
	return nil
}
