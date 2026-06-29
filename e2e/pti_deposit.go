package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iConnectAUSBankAccount links a US bank for the deposit/withdrawal setup flows.
// Plaid is now the only US bank-link path (PLAID_ENABLED=true in e2e), so this
// drives the Plaid overlay instead of the retired manual /connect/bank/us form.
// It links Tartan Bank checking — the deterministic mock bank — which is fine as
// setup because each scenario runs as a fresh user/wallet (no cross-scenario
// dedupe collision). The step name is unchanged so existing scenarios keep using
// it; only the mechanism moved to Plaid.
func (sc *E2EContext) iConnectAUSBankAccount() error {
	debugPrintln("\n🏦 Connecting US bank account via Plaid...")
	return sc.connectBankViaPlaid("Tartan Bank", "checking")
}

// iDepositViaPTIDepositForm fills the fynbos deposit form (amount + linked bank account)
// and completes the PTI confirm step at /deposit/:paymentId.
func (sc *E2EContext) iDepositViaPTIDepositForm(amount, currency string) error {
	debugPrintf("\n💵 Depositing %s %s via PTI deposit form...\n", amount, currency)

	// Fill the deposit amount field
	amountField := sc.page.Locator("#depositAmount, input[name='depositAmount']").First()
	if err := amountField.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("deposit amount field not found: %w", err)
	}
	if err := amountField.Fill(amount); err != nil {
		return fmt.Errorf("failed to fill deposit amount: %w", err)
	}
	debugPrintf("   ✓ Filled amount: %s\n", amount)

	if err := sc.iTakeAScreenshot("pti-deposit-form"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
	}

	// Click "Continue" to submit the deposit amount form
	continueBtn := sc.page.Locator("button[type='submit']:has-text('Continue'), button[form='account-deposit']").First()
	if err := continueBtn.Click(); err != nil {
		return fmt.Errorf("failed to click Continue on deposit form: %w", err)
	}

	// Wait for /deposit/:paymentId confirm page
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		currentURL := sc.page.URL()
		if strings.Contains(currentURL, "/deposit/") {
			debugPrintf("   ✓ Navigated to deposit confirm page: %s\n", currentURL)
			// Extract and store the payment/request ID from the URL
			parts := strings.Split(currentURL, "/deposit/")
			if len(parts) == 2 && parts[1] != "" {
				sc.ptiDepositRequestID = parts[1]
				debugPrintf("   ✓ Stored PTI deposit requestId: %s\n", sc.ptiDepositRequestID)
			}
			break
		}
		if i == 29 {
			return fmt.Errorf("did not navigate to deposit confirm page within 15 seconds")
		}
	}

	// Wait for page to settle
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})

	if err := sc.iTakeAScreenshot("pti-deposit-confirm"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
	}

	// Click "Confirm deposit"
	confirmBtn := sc.page.Locator("button:has-text('Confirm deposit'), button[form='deposit-confirm']").First()
	if err := confirmBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("Confirm deposit button not found: %w", err)
	}
	if err := confirmBtn.Click(); err != nil {
		return fmt.Errorf("failed to click Confirm deposit: %w", err)
	}

	// Wait for redirect back to dashboard
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		currentURL := sc.page.URL()
		if !strings.Contains(currentURL, "/deposit") {
			debugPrintf("   ✓ PTI deposit confirmed, navigated to: %s\n", currentURL)
			return nil
		}
	}

	return fmt.Errorf("PTI deposit did not complete within 30 seconds")
}

// iMockptiReturnsTheDeposit triggers a deposit return via the mockpti admin API.
// It POSTs feedback="RETURNED" to /transactions/{requestId}/updates and waits for
// the backend Temporal workflow to process the reversal.
func (sc *E2EContext) iMockptiReturnsTheDeposit() error {
	if sc.ptiDepositRequestID == "" {
		return fmt.Errorf("no PTI deposit requestId stored; deposit must be completed first")
	}

	debugPrintf("\n↩️  Triggering PTI deposit return for requestId: %s\n", sc.ptiDepositRequestID)

	payload := map[string]interface{}{
		"transactionId": sc.ptiDepositRequestID,
		"feedback":      "RETURNED",
		"providerName":  "test-provider",
		"payload":       `{"status":"RETURNED"}`,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal return payload: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	url := "https://mockpti.interledger.test/transactions/" + sc.ptiDepositRequestID + "/updates"
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create return request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-pti-client-id", "04d3e1b5-96d4-47e4-9eaa-13e9b4b0f219")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call mockpti return endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mockpti return endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	debugPrintf("   ✓ mockpti accepted deposit return, waiting for backend to process...\n")
	// Give the webhook + Temporal workflow time to process the reversal
	time.Sleep(3 * time.Second)
	return nil
}
