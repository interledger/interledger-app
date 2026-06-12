package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/playwright-community/playwright-go"
)

func (sc *E2EContext) iNavigateToTheWithdrawalPage() error {
	debugPrintln("\n💸 Navigating to withdrawal page...")

	url := sc.baseURL + "/withdraw"
	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to withdrawal page: %w", err)
	}

	// Wait for page to load
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Navigated to withdrawal page: %s\n", url)
	return nil
}

func (sc *E2EContext) iWithdrawViATheWithdrawalIframe(amount, currency string) error {
	debugPrintf("\n💸 Withdrawing %s %s via GateHub iframe...\n", amount, currency)

	// For EU users (Germany), the withdrawal page loads a GateHub iframe
	// NOT a frontend form. The iframe contains the amount/currency inputs.

	// Wait for page to be fully loaded
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	// Take screenshot of withdrawal page
	sc.iTakeAScreenshot("withdrawal-page")

	// Wait for the iframe to load
	debugPrintln("   Waiting for GateHub withdrawal iframe...")
	iframeLocator := sc.page.FrameLocator("iframe")

	// Look for amount input inside the iframe
	debugPrintln("   Looking for amount input in iframe...")
	amountInput := iframeLocator.Locator("input#amount").First()

	err := amountInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		sc.iTakeAScreenshot("withdrawal-iframe-not-found")
		return fmt.Errorf("amount input not found in withdrawal iframe: %w", err)
	}

	// Fill in the amount
	debugPrintf("   Filling amount: %s\n", amount)
	err = amountInput.Fill(amount)
	if err != nil {
		sc.iTakeAScreenshot("withdrawal-amount-fill-error")
		return fmt.Errorf("failed to fill amount field: %w", err)
	}
	debugPrintln("   ✓ Amount filled")

	// Retrieve the user's TOTP secret
	totpSecret, err := sc.getCurrentUserTOTPSecret()
	if err != nil {
		return fmt.Errorf("failed to get current user TOTP secret: %w", err)
	}

	// Look for the SCA TOTP Verification checkbox
	debugPrintln("   Looking for SCA trigger checkbox in iframe...")
	checkbox := iframeLocator.Locator("input#trigger_sca")

	err = checkbox.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		sc.iTakeAScreenshot("sca-checkbox-not-found")
		return fmt.Errorf("could not find sca trigger checkbox: %w", err)
	}

	// Set the checkbox to true
	err = checkbox.SetChecked(true)
	if err != nil {
		return fmt.Errorf("could not set sca trigger checkbox: %w", err)
	}
	debugPrintln("   ✓ SCA trigger checkbox set to true")

	// Look for the TOTP input field
	debugPrintln("   Looking for TOTP code input in iframe...")
	totpInput := iframeLocator.Locator("input#totp_code")

	err = totpInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		sc.iTakeAScreenshot("totp-input-not-found")
		return fmt.Errorf("could not find totp input: %w", err)
	}

	// Generate a TOTP code with the user's secret
	totpCode, err := generateTOTPCode(totpSecret)
	if err != nil {
		return fmt.Errorf("failed to generate TOTP code for SCA verification: %w", err)
	}
	debugPrintln("   ✓ Generated TOTP code")

	// Fill the TOTP input field
	err = totpInput.Fill(totpCode)
	if err != nil {
		sc.iTakeAScreenshot("totp-input-fill-error")
		return fmt.Errorf("failed to fill totp field: %w", err)
	}
	debugPrintln("   ✓ TOTP input filled")

	// Take screenshot showing filled form
	sc.iTakeAScreenshot("withdrawal-iframe-filled")

	// Click the "OK Complete" button using data-testid
	debugPrintln("   Clicking complete button in iframe...")
	completeButton := iframeLocator.Locator("[data-testid='complete-button']").First()

	err = completeButton.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("complete button not found: %w", err)
	}

	err = completeButton.Click()
	if err != nil {
		return fmt.Errorf("failed to click complete button: %w", err)
	}

	debugPrintln("   ✓ Complete button clicked in iframe")

	// Wait for withdrawal workflow to complete
	// The iframe sends postMessage → frontend catches → calls backend gRPC → workflow processes
	debugPrintln("   Waiting for withdrawal workflow to complete...")
	time.Sleep(5 * time.Second)

	debugPrintln("   ✓ Withdrawal processed")
	return nil
}

func (sc *E2EContext) listPendingGatehubWithdrawals(gatehubUserID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://mockgatehub.interledger.test/admin/users/%s/withdrawals?status=pending", gatehubUserID)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending withdrawals: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list pending withdrawals returned status %d: %s", resp.StatusCode, string(body))
	}

	var txs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return nil, fmt.Errorf("failed to decode pending withdrawals response: %w", err)
	}

	return txs, nil
}

func (sc *E2EContext) triggerGatehubWithdrawalEvent(txID, event string) error {
	url := fmt.Sprintf("https://mockgatehub.interledger.test/admin/withdrawals/%s/trigger-event", txID)

	payload, err := json.Marshal(map[string]string{"event": event})
	if err != nil {
		return fmt.Errorf("failed to marshal trigger event request: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create trigger event request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to trigger withdrawal event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("trigger withdrawal event returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (sc *E2EContext) theGatehubWithdrawalEventIsTriggered(event string) error {
	debugPrintf("\n💸 Triggering withdrawal event: %s\n", event)

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get current user email: %w", err)
	}

	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get GateHub user ID for user %s: %w", email, err)
	}

	txs, err := sc.listPendingGatehubWithdrawals(gatehubUserID)
	if err != nil {
		return fmt.Errorf("failed to list pending withdrawals: %w", err)
	}

	if len(txs) == 0 {
		return fmt.Errorf("no pending withdrawals found for user %s", gatehubUserID)
	}

	txID, ok := txs[0]["uuid"].(string)
	if !ok || txID == "" {
		return fmt.Errorf("pending withdrawal has no uuid field")
	}

	debugPrintf("   Triggering event %s for withdrawal %s\n", event, txID)

	if err := sc.triggerGatehubWithdrawalEvent(txID, event); err != nil {
		return fmt.Errorf("failed to trigger withdrawal event: %w", err)
	}

	// Wait for the webhook to be processed
	time.Sleep(5 * time.Second)

	debugPrintf("   ✓ Withdrawal event %s triggered\n", event)
	return nil
}

func (sc *E2EContext) thatGatehubChargesWithdrawalFee(feePercent string) error {
	debugPrintf("\n💸 Configuring GateHub withdrawal fee for user to %s%%...\n", feePercent)

	// Get the current user's email
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get current user email: %w", err)
	}

	// Get the GateHub managed user ID for this user
	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get GateHub user ID for user %s: %w", email, err)
	}

	// Parse fee percentage
	var feePct float64
	_, err = fmt.Sscanf(feePercent, "%f", &feePct)
	if err != nil {
		return fmt.Errorf("invalid fee percentage: %s", feePercent)
	}

	// Call MockGatehub's /admin/users/{userId}/fees endpoint
	url := fmt.Sprintf("https://mockgatehub.interledger.test/admin/users/%s/fees", gatehubUserID)
	payload := map[string]interface{}{
		"withdrawal_fee_percentage": feePct,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal fee config: %w", err)
	}

	// Use HTTP client for local testing
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create fee update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update withdrawal fee on MockGatehub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MockGatehub /admin/users/%s/fees returned status %d: %s", gatehubUserID, resp.StatusCode, string(body))
	}

	debugPrintf("✓ MockGatehub user-specific withdrawal fee configured to %s%% for user ID: %s\n", feePercent, gatehubUserID)
	return nil
}
