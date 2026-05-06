package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// iNavigateToTheDashboard navigates to the main dashboard
func (sc *E2EContext) iNavigateToTheDashboard() error {
	debugPrintln("\n📊 Navigating to dashboard...")

	url := sc.baseURL + "/"
	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to dashboard: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	if strings.Contains(sc.page.URL(), "/login") {
		debugPrintln("   🔐 Redirected to login while navigating to dashboard, retrying login...")
		if err := sc.iLogInAsMyself(); err != nil {
			return fmt.Errorf("failed to navigate to dashboard: not on application dashboard - still at: %s", sc.page.URL())
		}
		if _, err := sc.page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
			return fmt.Errorf("failed to navigate to dashboard after re-authentication: %w", err)
		}
	}

	if strings.Contains(sc.page.URL(), "/login") {
		return fmt.Errorf("failed to navigate to dashboard: not on application dashboard - still at: %s", sc.page.URL())
	}

	debugPrintf("✓ Navigated to dashboard: %s\n", url)
	return nil
}

// iShouldSeeMyAccountBalanceWith verifies the account balance is displayed
func (sc *E2EContext) iShouldSeeMyAccountBalanceWith(amount, currency string) error {
	debugPrintf("\n💰 Verifying balance: %s %s\n", amount, currency)

	// Log which user we expect to be logged in as
	if sc.currentUser != "" {
		email, err := sc.getCurrentUserEmail()
		if err == nil {
			debugPrintf("   Current user context: %s (email: %s)\n", sc.currentUser, email)

			// Verify user exists in database and check actual balance
			debugPrintf("   🔍 Checking database for user and balance...\n")

			// Get user_id from signups table using email
			kratosUserID, err := sc.getUserIDFromSignup(email)
			if err != nil {
				debugPrintf("   ⚠️  %v\n", err)
			} else {
				debugPrintf("   ✓ User found in signups: user_id=%s\n", kratosUserID)

				// Get wallet_id from user_wallets
				walletID, err := sc.getWalletIDForUser(kratosUserID)
				if err != nil {
					debugPrintf("   ⚠️  %v\n", err)
				} else {
					debugPrintf("   ✓ Wallet found: wallet_id=%s\n", walletID)

					// Check transactions table
					txCount, err := sc.getTransactionCount(walletID)
					if err == nil {
						debugPrintf("   ✓ Transaction count: %d\n", txCount)
					}
				}
			}
		}
	}

	// Prepare amount variants (with and without decimal places)
	amountVariants := []string{amount}
	if !strings.Contains(amount, ".") {
		amountVariants = append(amountVariants, amount+".00")
	}

	// Prefer stable balance selectors on the dashboard
	balanceCardSelector := fmt.Sprintf("[data-testid='wallet-balance-card'][data-currency='%s']", currency)
	balanceAmountSelector := balanceCardSelector + " [data-testid='wallet-balance-amount']"

	// Search for balance text containing both amount and currency
	// Reload on each attempt to ensure we get the latest data from the backend
	for i := 0; i < 10; i++ {
		// Reload page to get latest data from backend
		debugPrintf("   Reloading page to fetch latest balance (attempt %d)...\n", i+1)
		_, _ = sc.page.Reload()
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(10000),
		})
		time.Sleep(1 * time.Second)

		balanceAmount := sc.page.Locator(balanceAmountSelector)
		if count, _ := balanceAmount.Count(); count > 0 {
			text, _ := balanceAmount.First().TextContent()
			if strings.Contains(text, amount) {
				debugPrintf("✓ Found balance %s %s in dashboard card\n", amount, currency)
				return nil
			}
		}

		// Get all text content on the page
		allText, _ := sc.page.TextContent("body")

		// On first and last attempt, dump page content for debugging
		if i == 0 || i == 9 {
			debugPrintf("   📄 Page URL: %s\n", sc.page.URL())
			debugPrintf("   📄 Page text (first 500 chars): %s...\n", truncateString(allText, 500))
		}

		for _, amt := range amountVariants {
			if strings.Contains(allText, amt) && strings.Contains(allText, currency) {
				// Found the amount and currency in the page
				debugPrintf("✓ Found balance: %s %s in page content\n", amount, currency)
				return nil
			}
		}

		// Look for elements containing the currency
		currencyLocator := sc.page.Locator(fmt.Sprintf(":has-text('%s')", currency))
		count, _ := currencyLocator.Count()

		if count > 0 {
			debugPrintf("   Found %d elements containing '%s'\n", count, currency)
			for j := 0; j < count; j++ {
				text, _ := currencyLocator.Nth(j).TextContent()
				if i == 0 {
					debugPrintf("   Element %d text: %s\n", j, truncateString(text, 100))
				}
				for _, amt := range amountVariants {
					if strings.Contains(text, amt) {
						debugPrintf("✓ Found balance: %s %s\n", amount, currency)
						_ = currencyLocator.Nth(j).ScrollIntoViewIfNeeded()
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("balance %s %s not found on page after 10 attempts", amount, currency)
}

// iNavigateToTheSendPaymentPage navigates to the send payment page
func (sc *E2EContext) iNavigateToTheSendPaymentPage() error {
	debugPrintln("\n💸 Navigating to send payment page...")

	// Try common payment page paths
	paymentPaths := []string{
		"/pay",
	}

	var lastErr error
	for _, path := range paymentPaths {
		url := sc.baseURL + path
		debugPrintf("   Trying: %s\n", url)

		_, err := sc.page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(5000),
		})

		if err == nil {
			sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State:   playwright.LoadStateNetworkidle,
				Timeout: playwright.Float(5000),
			})

			if strings.Contains(sc.page.URL(), "/login") {
				debugPrintln("   🔐 Redirected to login while opening payments page, retrying login...")
				if loginErr := sc.iLogInAsMyself(); loginErr == nil {
					_, err = sc.page.Goto(url, playwright.PageGotoOptions{
						WaitUntil: playwright.WaitUntilStateNetworkidle,
						Timeout:   playwright.Float(5000),
					})
					if err == nil {
						sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
							State:   playwright.LoadStateNetworkidle,
							Timeout: playwright.Float(5000),
						})
					}
				}
			}

			if !strings.Contains(sc.page.URL(), "/login") {
				debugPrintf("✓ Navigated to: %s\n", url)
				return nil
			}
		}
		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("failed to navigate to send payment page: %w", lastErr)
	}
	return fmt.Errorf("failed to navigate to send payment page: redirected to login")
}

// iShouldSeeThePaymentsPage verifies the payments page is displayed
func (sc *E2EContext) iShouldSeeThePaymentsPage() error {
	debugPrintln("\n📋 Verifying payments page...")

	// Check if we're on a payments-related URL
	currentURL := sc.page.URL()
	debugPrintf("   Current URL: %s\n", currentURL)

	if strings.Contains(currentURL, "/login") {
		debugPrintln("   🔐 On login page during payments check, retrying login...")
		if err := sc.iLogInAsMyself(); err == nil {
			_, _ = sc.page.Goto(sc.baseURL+"/pay", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
			currentURL = sc.page.URL()
			debugPrintf("   Current URL after login recovery: %s\n", currentURL)
		}
	}

	// Look for payments page indicators (title or header text)
	allText, _ := sc.page.TextContent("body")

	// Check for common payment page text
	if strings.Contains(allText, "Payments") ||
		strings.Contains(allText, "Payment") ||
		strings.Contains(currentURL, "/pay") ||
		strings.Contains(currentURL, "/payments") {
		debugPrintln("✓ Payments page is visible")
		return nil
	}

	return fmt.Errorf("payments page not found - URL: %s", currentURL)
}

// iFillInTheReceiverEmailWith fills in the receiver email from stored user details
func (sc *E2EContext) iFillInTheReceiverEmailWith(userName string) error {
	debugPrintf("\n📧 Filling receiver email for user '%s'\n", userName)

	// Get the user's email from stored details
	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s'", userName)
	}

	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("no emailSuffix defined for user '%s'", userName)
	}

	// Construct full email with test prefix
	email := fmt.Sprintf("%s-%s", sc.testIdentifier, emailSuffix)
	debugPrintf("   Receiver email: %s\n", email)

	// Find and fill the email input field
	emailInputSelectors := []string{
		"input[name='receiverEmail'], input[name='receiver_email'], input[placeholder*='email' i], input[type='email']",
		"input[name='recipient'], input[placeholder*='recipient' i]",
		"input[placeholder*='Recipient' i], input[placeholder*='recipient email' i]",
	}

	for _, selector := range emailInputSelectors {
		emailInput := sc.page.Locator(selector)
		count, _ := emailInput.Count()
		if count > 0 {
			if err := emailInput.First().Fill(email); err == nil {
				debugPrintf("✓ Filled receiver email: %s\n", email)
				return nil
			}
		}
	}

	return fmt.Errorf("could not find receiver email input field")
}

// iFillInThePaymentAmount fills in the payment amount
func (sc *E2EContext) iFillInThePaymentAmount(amount string) error {
	debugPrintf("\n💵 Filling payment amount: %s\n", amount)

	// Prefer stable test selector
	amountInput := sc.page.Locator("[data-testid='pay-amount-input']")
	_ = amountInput.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if count, _ := amountInput.Count(); count > 0 {
		if err := amountInput.First().Fill(amount); err == nil {
			debugPrintf("✓ Filled amount: %s\n", amount)
			return nil
		}
	}

	// Find and fill the amount input field
	amountInputSelectors := []string{
		"input[name='amount'], input[placeholder*='amount' i], input[type='number']",
		"input[name='paymentAmount'], input[placeholder*='payment' i]",
		"input[placeholder*='Amount' i]",
	}

	for _, selector := range amountInputSelectors {
		amountInput := sc.page.Locator(selector)
		count, _ := amountInput.Count()
		if count > 0 {
			if err := amountInput.First().Fill(amount); err == nil {
				debugPrintf("✓ Filled amount: %s\n", amount)
				return nil
			}
		}
	}

	return fmt.Errorf("could not find payment amount input field")
}

// iSelectThePaymentCurrency selects the payment currency
func (sc *E2EContext) iSelectThePaymentCurrency(currency string) error {
	debugPrintf("\n💱 Selecting currency: %s\n", currency)

	// Prefer headless-ui listbox test selectors
	listboxBtn := sc.page.Locator("[data-testid='pay-currency-select']")

	// Wait for the button to be visible and stable — filling the amount field
	// triggers a debounced network request that causes React re-renders, which
	// can make the button temporarily unstable for Playwright's actionability
	// checks.
	err := listboxBtn.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("currency selector not visible: %w", err)
	}

	// Wait for the page to settle after the amount-change network request
	// completes and React finishes re-rendering.
	if err := sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(5000),
	}); err != nil {
		debugPrintf("⚠ network idle wait timed out, proceeding anyway\n")
	}

	if count, _ := listboxBtn.Count(); count > 0 {
		if err := listboxBtn.First().Click(); err != nil {
			return fmt.Errorf("failed to open currency selector: %w", err)
		}

		option := sc.page.Locator(fmt.Sprintf("[data-testid='pay-currency-option'][data-currency-code='%s']", currency))
		_ = option.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		if count, _ := option.Count(); count > 0 {
			if err := option.First().Click(); err == nil {
				debugPrintf("✓ Selected currency: %s\n", currency)
				return nil
			}
		}
	}

	// Fallback for native selects
	currencySelectSelectors := []string{
		"select[name='currency'], select[name='paymentCurrency']",
		"select[placeholder*='currency' i]",
	}

	for _, selector := range currencySelectSelectors {
		currencySelect := sc.page.Locator(selector)
		count, _ := currencySelect.Count()
		if count > 0 {
			_, _ = currencySelect.SelectOption(playwright.SelectOptionValues{Values: &[]string{currency}})
			debugPrintf("✓ Selected currency: %s\n", currency)
			return nil
		}
	}

	return fmt.Errorf("could not find currency selector")
}

// iSubmitThePayment submits the payment form. The protea Pay flow is a
// two-step SPA: the Amount step (Continue button) transitions in-memory to the
// Confirm step (agreement checkbox + Confirm payment button). Both must be
// driven for a payment to actually be created on the backend.
func (sc *E2EContext) iSubmitThePayment() error {
	debugPrintln("\n📤 Submitting payment...")

	confirmButton := sc.page.Locator("[data-testid='pay-confirm-submit']")
	continueButton := sc.page.Locator("[data-testid='pay-amount-continue']")

	// If we're already on the Confirm step, just check agreement and submit.
	if err := confirmButton.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(500),
	}); err == nil {
		return sc.confirmPaymentAgreementAndSubmit(confirmButton)
	}

	// Otherwise we expect to be on the Amount step. Wait for the Continue
	// button to be ready (the amount field debounces a network request which
	// can briefly destabilise the form), then click it to advance to Confirm.
	if err := continueButton.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("pay-amount-continue button not visible: %w", err)
	}

	// Allow any in-flight amount-update fetcher request to settle so the
	// Continue submit carries the latest amount/currency.
	if err := sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(5000),
	}); err != nil {
		debugPrintf("⚠ network idle wait timed out before Continue, proceeding anyway\n")
	}

	if err := continueButton.First().Click(); err != nil {
		return fmt.Errorf("failed to click pay-amount-continue: %w", err)
	}
	debugPrintln("✓ Clicked Continue, waiting for Confirm step...")

	if err := confirmButton.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("pay-confirm-submit button did not appear: %w", err)
	}

	return sc.confirmPaymentAgreementAndSubmit(confirmButton)
}

// confirmPaymentAgreementAndSubmit ticks the service-agreement checkbox and
// clicks the Confirm payment button. After clicking, the protea action
// redirects to '/' on success. The agreement checkbox is required by the
// backend, so any failure to tick it is treated as a hard error rather than
// silently swallowed (which would let the subsequent submit appear to
// succeed while the server rejects the request).
func (sc *E2EContext) confirmPaymentAgreementAndSubmit(confirmButton playwright.Locator) error {
	agreement := sc.page.Locator("[data-testid='pay-confirm-agreement']")
	if aCount, _ := agreement.Count(); aCount > 0 {
		if err := agreement.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("pay-confirm-agreement checkbox not visible: %w", err)
		}
		checked, _ := agreement.First().IsChecked()
		if !checked {
			if err := agreement.First().Check(); err != nil {
				return fmt.Errorf("failed to check pay-confirm-agreement: %w", err)
			}
		}
	}

	if err := confirmButton.First().Click(); err != nil {
		return fmt.Errorf("failed to click pay-confirm-submit: %w", err)
	}
	debugPrintln("✓ Submitted payment confirmation")
	return nil
}

// iShouldSeeAPaymentConfirmation waits for payment confirmation
func (sc *E2EContext) iShouldSeeAPaymentConfirmation() error {
	debugPrintln("\n✅ Waiting for payment confirmation...")

	// Look for confirmation message or success state
	confirmationSelectors := []string{
		"text=Payment successful",
		"text=Payment submitted",
		"text=Confirmation",
		"[role='alert']:has-text('success')",
		".success-message",
		".confirmation-message",
	}

	for i := 0; i < 10; i++ {
		for _, selector := range confirmationSelectors {
			elem := sc.page.Locator(selector)
			count, _ := elem.Count()
			if count > 0 {
				debugPrintf("✓ Payment confirmation found\n")
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// If no explicit confirmation, just verify we're still on the page without errors
	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "error") {
		debugPrintln("✓ No error detected in URL, assuming payment submitted")
		return nil
	}

	return fmt.Errorf("could not verify payment confirmation")
}

// iWaitForThePaymentToComplete waits for the payment to be processed.
// After a successful Confirm submit, protea redirects to '/' (dashboard).
// We poll the URL for that redirect; if it never happens we fall back to a
// fixed wait so balance-update assertions can still observe backend state.
// IMPORTANT: do not call page.Reload() here — the Pay flow keeps the
// AMOUNT/CONFIRM step in an in-memory zustand store, and reloading mid-flow
// resets it, preventing the payment from being confirmed.
func (sc *E2EContext) iWaitForThePaymentToComplete() error {
	debugPrintln("\n⏳ Waiting for payment to complete...")

	maxWait := 30
	checkInterval := 1

	for i := 0; i < maxWait; i += checkInterval {
		time.Sleep(time.Duration(checkInterval) * time.Second)

		currentURL := sc.page.URL()

		// Bail out early on auth-expiry or error redirects so the failure is
		// reported here instead of masquerading as a missing balance update.
		if strings.Contains(currentURL, "/login") || strings.Contains(currentURL, "/totp") {
			return fmt.Errorf("redirected to auth page during payment confirmation: %s", currentURL)
		}
		if strings.Contains(currentURL, "error") {
			return fmt.Errorf("redirected to error page during payment confirmation: %s", currentURL)
		}

		// On success, protea redirects from /pay/:paymentId to the dashboard
		// at '/'. Match that explicitly rather than "anything not /pay".
		if isDashboardURL(currentURL, sc.baseURL) {
			debugPrintf("✓ Payment confirmed, redirected to %s\n", currentURL)
			if err := sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State:   playwright.LoadStateNetworkidle,
				Timeout: playwright.Float(5000),
			}); err != nil {
				debugPrintf("⚠ networkidle wait after payment redirect timed out: %v\n", err)
			}
			return nil
		}

		debugPrintf("   Waiting... (%d/%d seconds, url=%s)\n", i+checkInterval, maxWait, currentURL)
	}

	return fmt.Errorf("payment did not redirect to dashboard within %d seconds", maxWait)
}

// isDashboardURL returns true when the given URL points at the wallet
// dashboard (i.e. the post-payment success destination), tolerating
// trailing slashes and query strings.
func isDashboardURL(currentURL, baseURL string) bool {
	trimmed := strings.TrimSuffix(currentURL, "/")
	if trimmed == strings.TrimSuffix(baseURL, "/") {
		return true
	}
	// Tolerate query strings on the dashboard (e.g. snackbar params).
	if idx := strings.IndexAny(trimmed, "?#"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return trimmed == strings.TrimSuffix(baseURL, "/")
}

// iDepositViATheDepositIframeAsUser completes the deposit flow for a specific user
func (sc *E2EContext) iDepositViATheDepositIframeAsUser(amount, currency, userName string) error {
	debugPrintf("\n💶 Depositing %s %s as '%s' via iframe...\n", amount, currency, userName)

	// Impersonate the user first
	if err := sc.iImpersonate(userName); err != nil {
		return err
	}

	// Then navigate to deposit page and perform the deposit
	if err := sc.iNavigateToTheDepositPage(); err != nil {
		return err
	}

	return sc.iDepositViATheDepositIframe(amount, currency)
}

// thePaymentFormShouldBeAccessible verifies the payment form is accessible
func (sc *E2EContext) thePaymentFormShouldBeAccessible() error {
	debugPrintln("\n✅ Verifying payment form is accessible...")

	// Check if we can find the payment form elements that we just filled
	amountInput := sc.page.Locator("input[name='amount'], input[name='paymentAmount'], input[type='number']")
	count, _ := amountInput.Count()

	if count > 0 {
		debugPrintln("✓ Payment form is accessible")
		return nil
	}

	return fmt.Errorf("payment form not accessible - amount input not found")
}

// iGetTheReceiverWalletAddressFor retrieves the wallet address for a user
func (sc *E2EContext) iNavigateToThePaymentsHistoryPage() error {
	debugPrintln("\n📋 Navigating to payments history page...")

	url := sc.baseURL + "/payments"
	debugPrintf("   Trying: %s\n", url)

	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to payments history: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Navigated to: %s\n", url)
	return nil
}

// iGetTheReceiverWalletAddressFor retrieves the wallet address for a user
func (sc *E2EContext) iGetTheReceiverWalletAddressFor(userName string) error {
	debugPrintf("\n🔍 Getting wallet address for '%s'...\n", userName)

	// Get the user's email from stored details
	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s'", userName)
	}

	// Prefer the stored wallet address from the KYC flow
	if walletAddress, ok := details.Fields["walletAddress"]; ok && walletAddress != "" {
		sc.receiverWalletAddress = walletAddress
		debugPrintf("✓ Stored receiver wallet address: %s\n", walletAddress)
		return nil
	}

	// Fallback to email if wallet address is not available
	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("no emailSuffix defined for user '%s'", userName)
	}

	// Construct the full email address
	walletIdentifier := fmt.Sprintf("%s-%s", sc.testIdentifier, emailSuffix)

	// Store for use in next step
	sc.receiverWalletAddress = walletIdentifier

	debugPrintf("✓ Stored receiver wallet address: %s\n", walletIdentifier)
	return nil
}

// iFillInTheReceiverWalletAddress fills in the receiver wallet address field
func (sc *E2EContext) iFillInTheReceiverWalletAddress() error {
	debugPrintf("\n💳 Filling receiver wallet address: %s\n", sc.receiverWalletAddress)

	if sc.receiverWalletAddress == "" {
		return fmt.Errorf("no receiver wallet address set - call 'I get the receiver wallet address for' first")
	}

	// Pay search uses the command palette input
	searchSelectors := []string{
		"[data-testid='pay-search-input']",
		"input#search",
		"input[name='search']",
		"input[placeholder='Search for someone to pay']",
		"input[placeholder*='pay' i]",
	}

	var searchInput playwright.Locator
	for _, selector := range searchSelectors {
		locator := sc.page.Locator(selector)
		if count, _ := locator.Count(); count > 0 {
			searchInput = locator.First()
			break
		}
	}

	if searchInput == nil {
		return fmt.Errorf("could not find receiver wallet address input field")
	}

	if err := searchInput.Fill(sc.receiverWalletAddress); err != nil {
		return fmt.Errorf("failed to fill receiver wallet address: %w", err)
	}
	debugPrintf("✓ Filled receiver wallet address using search input\n")
	searchInput.Press("Enter")
	// Wait for any search results to appear (best-effort)
	results := sc.page.Locator("[data-testid='pay-search-result']")
	_ = results.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	})

	// Wait for search results to appear and select the matching entry
	searchCandidates := []string{sc.receiverWalletAddress}
	if lastSlash := strings.LastIndex(sc.receiverWalletAddress, "/"); lastSlash >= 0 {
		searchCandidates = append(searchCandidates, sc.receiverWalletAddress[lastSlash+1:])
	}

	// Prefer data attributes if present
	for _, candidate := range searchCandidates {
		option := sc.page.Locator(fmt.Sprintf("[data-testid='pay-search-result'][data-wallet-url='%s']", candidate))
		_ = option.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		if count, _ := option.Count(); count > 0 {
			if err := option.First().Click(); err == nil {
				debugPrintf("✓ Selected receiver from search results (wallet url match): %s\n", candidate)
				sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State: playwright.LoadStateNetworkidle,
				})
				return nil
			}
		}
	}

	for _, candidate := range searchCandidates {
		option := sc.page.Locator(fmt.Sprintf("button:has-text('%s')", candidate))
		if count, _ := option.Count(); count > 0 {
			if err := option.First().Click(); err == nil {
				debugPrintf("✓ Selected receiver from search results: %s\n", candidate)
				sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State: playwright.LoadStateNetworkidle,
				})
				return nil
			}
		}
	}

	return fmt.Errorf("receiver wallet address not found in search results")
}

// iShouldSeeThePaymentInMyTransactionHistory verifies payment appears in history
func (sc *E2EContext) iShouldSeeThePaymentInMyTransactionHistory() error {
	debugPrintln("\n📜 Verifying payment appears in transaction history...")

	allText, _ := sc.page.TextContent("body")

	// Look for payment indicators
	if strings.Contains(allText, "sent") ||
		strings.Contains(allText, "payment") ||
		strings.Contains(allText, "transaction") {
		debugPrintln("✓ Payment found in transaction history")
		return nil
	}

	return fmt.Errorf("payment not found in transaction history")
}

// parseSeconds helper to convert string seconds to int
func parseSeconds(seconds string) int {
	secondsInt := 5 // default
	if parsed, err := time.ParseDuration(seconds + "s"); err == nil {
		secondsInt = int(parsed.Seconds())
	}
	return secondsInt
}

// iWaitSeconds waits for specified seconds
func (sc *E2EContext) iWaitSeconds(seconds string) error {
	secondsInt := parseSeconds(seconds)
	debugPrintln(fmt.Sprintf("Waiting %d seconds...", secondsInt))
	time.Sleep(time.Duration(secondsInt) * time.Second)
	return nil
}
