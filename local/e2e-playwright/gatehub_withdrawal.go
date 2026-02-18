package main

import (
	"bytes"
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

func (sc *E2EContext) thatGatehubChargesWithdrawalFee(feePercent string) error {
	debugPrintf("\n💸 Configuring Gatehub withdrawal fee to %s%%...\n", feePercent)

	// Parse fee percentage
	var feePct float64
	_, err := fmt.Sscanf(feePercent, "%f", &feePct)
	if err != nil {
		return fmt.Errorf("invalid fee percentage: %s", feePercent)
	}

	// Call MockGatehub's /admin/fees endpoint
	url := "https://mockgatehub.interledger.test/admin/fees"
	payload := map[string]interface{}{
		"withdrawal_fee_percentage": feePct,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal fee config: %w", err)
	}

	// Use HTTP client for local testing
	httpClient := &http.Client{Timeout: 10 * time.Second}

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
		return fmt.Errorf("MockGatehub /admin/fees returned status %d: %s", resp.StatusCode, string(body))
	}

	debugPrintf("✓ MockGatehub withdrawal fee configured to %s%%\n", feePercent)
	return nil
}
