package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iNavigateToTheDepositPage navigates to the deposit page
func (sc *E2EContext) iNavigateToTheDepositPage() error {
	debugPrintln("\n💰 Navigating to deposit page...")

	// Check if we're on login page - if so, the user session was lost during KYC
	currentURL := sc.page.URL()
	if strings.Contains(currentURL, "/login") {
		return fmt.Errorf("cannot navigate to deposit page: user is on login page - session was lost during KYC completion. Current URL: %s", currentURL)
	}

	url := sc.baseURL + "/deposit"
	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to deposit page: %w", err)
	}

	// Wait for page to load
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Navigated to deposit page: %s\n", url)
	return nil
}

// iDepositViATheDepositIframe completes the entire deposit flow
func (sc *E2EContext) iDepositViATheDepositIframe(amount, currency string) error {
	debugPrintf("\n💶 Depositing %s %s via iframe...\n", amount, currency)

	// Wait for deposit iframe to load
	debugPrintln("   Waiting for deposit iframe...")
	iframeLocator := sc.page.Locator("iframe[src*='mockgatehub'], iframe[title*='deposit' i], iframe")
	frameLocator := sc.page.FrameLocator("iframe[src*='mockgatehub'], iframe[title*='deposit' i], iframe")

	// Wait for iframe to be present
	for i := 0; i < 50; i++ {
		count, _ := iframeLocator.Count()
		if count > 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Set up a listener to capture the deposit completion postMessage
	_, err := sc.page.Evaluate("() => { window.depositCompleted = false; window.depositError = null; window.addEventListener('message', function(e){ var data = (e && e.data) ? e.data : {}; if ((data.type === 'deposit-complete' && data.status === 'success') || data.type === 'StripeDepositCompleted') { window.depositCompleted = true; } if (data.type === 'deposit-complete' && data.status === 'error') { window.depositError = data.message || \"deposit error\"; } }); }")
	if err != nil {
		debugPrintf("   ⚠️  Failed to set up deposit message listener: %v\n", err)
	}

	// Fill in the deposit form within iframe
	debugPrintln("   Filling deposit form...")

	// Fill amount (MockGatehub iframe uses #amount)
	amountField := frameLocator.Locator("#amount")
	if err := amountField.First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("could not find amount field: %w", err)
	}
	if err := amountField.First().Fill(amount); err != nil {
		return fmt.Errorf("failed to fill amount: %w", err)
	}
	debugPrintf("   ✓ Filled amount: %s\n", amount)

	// Select currency after options load
	currencySelect := frameLocator.Locator("#currency")
	if err := currencySelect.First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("could not find currency selector: %w", err)
	}
	// Wait for options to populate
	for i := 0; i < 50; i++ {
		optionCount, _ := currencySelect.Locator("option").Count()
		if optionCount > 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	_, _ = currencySelect.SelectOption(playwright.SelectOptionValues{Values: &[]string{currency}})
	debugPrintf("   ✓ Selected currency: %s\n", currency)

	// Submit the form
	submitSelectors := []string{
		"button:has-text('OK Complete')",
		"button:has-text('Complete')",
		"button.button.success",
		"button[type='submit']",
		"button:has-text('Submit')",
		"button:has-text('Deposit')",
		"button:has-text('Confirm')",
		"input[type='submit']",
	}

	submitted := false
	for _, selector := range submitSelectors {
		submitBtn := frameLocator.Locator(selector)
		count, err := submitBtn.Count()
		if err == nil && count > 0 {
			if err := submitBtn.First().Click(); err == nil {
				debugPrintln("   ✓ Submitted deposit form")
				submitted = true
				break
			}
		}
	}

	if !submitted {
		return fmt.Errorf("could not find or click submit button")
	}

	// Wait for deposit completion message from iframe (mockgatehub sends postMessage)
	debugPrintln("   Waiting for deposit completion message...")
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(500 * time.Millisecond)
		if errVal, _ := sc.page.Evaluate(`() => window.depositError`); errVal != nil {
			if errStr, ok := errVal.(string); ok && errStr != "" {
				return fmt.Errorf("deposit completion error: %s", errStr)
			}
		}
		completed, _ := sc.page.Evaluate(`() => window.depositCompleted === true`)
		if completed == true {
			debugPrintf("   ✓ Deposit completion message received (attempt %d/%d)\n", i+1, maxAttempts)
			return nil
		}
		if i%10 == 0 {
			debugPrintf("   ... still waiting (%d/%d)\n", i+1, maxAttempts)
		}
	}

	return fmt.Errorf("deposit completion message not received within %d seconds", maxAttempts/2)
}

// thatGatehubChargesDepositFee configures MockGatehub to charge a deposit fee for the current user
func (sc *E2EContext) thatGatehubChargesDepositFee(feePercent string) error {
	debugPrintf("\n💰 Configuring GateHub deposit fee for user to %s%%...\n", feePercent)

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

	// Convert feePercent string to float64
	feePct, err := strconv.ParseFloat(feePercent, 64)
	if err != nil {
		return fmt.Errorf("failed to parse fee percentage: %w", err)
	}

	// Call MockGatehub's /admin/users/{userId}/fees endpoint to set user-specific deposit fee
	mockgatehubURL := "https://mockgatehub.interledger.test"
	feeEndpoint := fmt.Sprintf("%s/admin/users/%s/fees", mockgatehubURL, gatehubUserID)

	// Create the request payload with numeric fee value
	payload := map[string]interface{}{
		"deposit_fee_percentage": feePct,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal fee config: %w", err)
	}

	// Create and send the PUT request
	req, err := http.NewRequest("PUT", feeEndpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create fee config request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Use a client with InsecureSkipVerify for local testing
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set fee config on MockGatehub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MockGatehub /admin/users/%s/fees returned status %d: %s", gatehubUserID, resp.StatusCode, string(body))
	}

	debugPrintf("✓ MockGatehub user-specific deposit fee configured to %s%% for user ID: %s\n", feePercent, gatehubUserID)
	return nil
}
