package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iNavigateToTheDepositPage navigates to the deposit page
func (sc *SignupContext) iNavigateToTheDepositPage() error {
	debugPrintln("\n💰 Navigating to deposit page...")

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
func (sc *SignupContext) iDepositViATheDepositIframe(amount, currency string) error {
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
	_, err := sc.page.Evaluate("() => { window.depositCompleted = false; window.depositError = null; window.addEventListener('message', function(e){ var data = (e && e.data) ? e.data : {}; if (data.type === 'deposit-complete' && data.status === 'success') { window.depositCompleted = true; } if (data.type === 'deposit-complete' && data.status === 'error') { window.depositError = data.message || \"deposit error\"; } }); }")
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

// iShouldSeeMyBalanceUpdatedWithAmount verifies the balance was updated
func (sc *SignupContext) iShouldSeeMyBalanceUpdatedWithAmount(amount, currency string) error {
	debugPrintf("\n💰 Verifying balance updated with %s %s...\n", amount, currency)

	// Navigate to dashboard to check updated balance UI
	_, _ = sc.page.Goto(sc.baseURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})

	// Allow time for webhook processing and UI refresh
	amountVariants := []string{amount}
	if !strings.Contains(amount, ".") {
		amountVariants = append(amountVariants, amount+".00")
	}

	findBalanceMatch := func() bool {
		balanceLocator := sc.page.Locator(":has-text('Balance'), :has-text('balance')")
		count, _ := balanceLocator.Count()
		if count > 20 {
			count = 20
		}
		for j := 0; j < count; j++ {
			text, _ := balanceLocator.Nth(j).TextContent()
			for _, amt := range amountVariants {
				if strings.Contains(text, currency) && strings.Contains(text, amt) {
					_ = balanceLocator.Nth(j).ScrollIntoViewIfNeeded()
					return true
				}
			}
		}

		currencyLocator := sc.page.Locator(fmt.Sprintf(":has-text(\"%s\")", currency))
		count, _ = currencyLocator.Count()
		if count > 20 {
			count = 20
		}
		for j := 0; j < count; j++ {
			text, _ := currencyLocator.Nth(j).TextContent()
			for _, amt := range amountVariants {
				if strings.Contains(text, amt) {
					_ = currencyLocator.Nth(j).ScrollIntoViewIfNeeded()
					return true
				}
			}
		}

		return false
	}

	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		_, _ = sc.page.Reload()

		if findBalanceMatch() {
			// Force one more refresh before capturing the screenshot
			_, _ = sc.page.Reload()
			if findBalanceMatch() {
				debugPrintf("✓ Balance appears updated on UI (attempt %d)\n", i+1)
				_ = sc.iTakeAScreenshot("deposit-balance-updated")
				return nil
			}
		}
	}

	return fmt.Errorf("balance update not visible on UI after waiting")
}
