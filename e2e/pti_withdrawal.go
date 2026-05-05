package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iWithdrawViaPTIWithdrawForm fills the fynbos withdrawal form (amount + linked bank account)
// and completes the PTI confirm step at /withdraw/:paymentId.
func (sc *E2EContext) iWithdrawViaPTIWithdrawForm(amount, currency string) error {
	debugPrintf("\n💸 Withdrawing %s %s via PTI withdraw form...\n", amount, currency)

	// Fill the withdrawal amount field
	amountField := sc.page.Locator("#withdrawAmount, input[name='withdrawAmount']").First()
	if err := amountField.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("withdrawal amount field not found: %w", err)
	}
	if err := amountField.Fill(amount); err != nil {
		return fmt.Errorf("failed to fill withdrawal amount: %w", err)
	}
	debugPrintf("   ✓ Filled amount: %s\n", amount)

	if err := sc.iTakeAScreenshot("pti-withdrawal-form"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
	}

	// Click "Continue" to submit the withdrawal amount form
	continueBtn := sc.page.Locator("button[type='submit']:has-text('Continue'), button[form='account-withdraw']").First()
	if err := continueBtn.Click(); err != nil {
		return fmt.Errorf("failed to click Continue on withdrawal form: %w", err)
	}

	// Wait for /withdraw/:paymentId confirm page
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		currentURL := sc.page.URL()
		if strings.Contains(currentURL, "/withdraw/") {
			debugPrintf("   ✓ Navigated to withdrawal confirm page: %s\n", currentURL)
			break
		}
		if i == 29 {
			return fmt.Errorf("did not navigate to withdrawal confirm page within 15 seconds")
		}
	}

	// Wait for page to settle
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	if err := sc.iTakeAScreenshot("pti-withdrawal-confirm"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
	}

	// Click "Confirm withdraw"
	confirmBtn := sc.page.Locator("button:has-text('Confirm withdraw'), button[form='withdraw-confirm']").First()
	if err := confirmBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("Confirm withdraw button not found: %w", err)
	}
	if err := confirmBtn.Click(); err != nil {
		return fmt.Errorf("failed to click Confirm withdraw: %w", err)
	}

	// Wait for redirect back to dashboard
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		currentURL := sc.page.URL()
		if !strings.Contains(currentURL, "/withdraw") {
			debugPrintf("   ✓ PTI withdrawal confirmed, navigated to: %s\n", currentURL)
			return nil
		}
	}

	return fmt.Errorf("PTI withdrawal did not complete within 30 seconds")
}
