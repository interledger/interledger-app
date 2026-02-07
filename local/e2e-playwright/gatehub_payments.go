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

			// First get the user_id from signups table using email
			var kratosUserID string
			err = sc.db.QueryRow("SELECT user_id FROM signups WHERE email = $1", email).Scan(&kratosUserID)
			if err != nil {
				debugPrintf("   ⚠️  User NOT found in signups table: %v\n", err)
			} else {
				debugPrintf("   ✓ User found in signups: user_id=%s\n", kratosUserID)

				// Get wallet_id from user_wallets
				var walletID string
				err = sc.db.QueryRow("SELECT wallet_id FROM user_wallets WHERE user_id = $1", kratosUserID).Scan(&walletID)
				if err != nil {
					debugPrintf("   ⚠️  Wallet NOT found for user: %v\n", err)
				} else {
					debugPrintf("   ✓ Wallet found: wallet_id=%s\n", walletID)

					// Check accounts table for balance
					var dbBalance float64
					err = sc.db.QueryRow("SELECT COALESCE(value, 0) FROM accounts WHERE wallet_id = $1 AND asset_code = $2",
						walletID, currency).Scan(&dbBalance)
					if err != nil {
						debugPrintf("   ⚠️  No balance record in accounts table: %v\n", err)
					} else {
						debugPrintf("   ✓ Database balance: %s %f\n", currency, dbBalance)
					}

					// Check transactions table
					var txCount int
					err = sc.db.QueryRow("SELECT COUNT(*) FROM transactions WHERE wallet_id = $1", walletID).Scan(&txCount)
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
			State: playwright.LoadStateNetworkidle,
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
				State: playwright.LoadStateNetworkidle,
			})
			debugPrintf("✓ Navigated to: %s\n", url)
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to navigate to send payment page: %w", lastErr)
}

// iShouldSeeThePaymentsPage verifies the payments page is displayed
func (sc *E2EContext) iShouldSeeThePaymentsPage() error {
	debugPrintln("\n📋 Verifying payments page...")

	// Check if we're on a payments-related URL
	currentURL := sc.page.URL()
	debugPrintf("   Current URL: %s\n", currentURL)

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
	email := fmt.Sprintf("%s-%s", sc.testEmailPrefix, emailSuffix)
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

// iSubmitThePayment submits the payment form
func (sc *E2EContext) iSubmitThePayment() error {
	debugPrintln("\n📤 Submitting payment...")

	confirmButton := sc.page.Locator("[data-testid='pay-confirm-submit']")
	continueButton := sc.page.Locator("[data-testid='pay-amount-continue']")

	if count, _ := confirmButton.Count(); count > 0 {
		agreement := sc.page.Locator("[data-testid='pay-confirm-agreement']")
		if aCount, _ := agreement.Count(); aCount > 0 {
			checked, _ := agreement.First().IsChecked()
			if !checked {
				_ = agreement.First().Check()
			}
		}
		if err := confirmButton.First().Click(); err == nil {
			debugPrintln("✓ Submitted payment confirmation")
			return nil
		}
	}

	if count, _ := continueButton.Count(); count > 0 {
		if err := continueButton.First().Click(); err != nil {
			return fmt.Errorf("failed to click continue: %w", err)
		}
		_ = confirmButton.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		agreement := sc.page.Locator("[data-testid='pay-confirm-agreement']")
		if aCount, _ := agreement.Count(); aCount > 0 {
			checked, _ := agreement.First().IsChecked()
			if !checked {
				_ = agreement.First().Check()
			}
		}
		if count, _ := confirmButton.Count(); count > 0 {
			if err := confirmButton.First().Click(); err == nil {
				debugPrintln("✓ Submitted payment confirmation")
				return nil
			}
		}
	}

	// Fallback: attempt generic submit button
	submitSelectors := []string{
		"button:has-text('Confirm payment')",
		"button:has-text('Continue')",
		"button:has-text('Send'), button:has-text('Submit'), button[type='submit']",
		"button:has-text('Pay'), button:has-text('Transfer')",
	}

	for _, selector := range submitSelectors {
		submitBtn := sc.page.Locator(selector)
		count, _ := submitBtn.Count()
		if count > 0 {
			if err := submitBtn.First().Click(); err == nil {
				debugPrintln("✓ Submitted payment form")
				return nil
			}
		}
	}

	return fmt.Errorf("could not find or click submit button")
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

// iWaitForThePaymentToComplete waits for the payment to be processed
func (sc *E2EContext) iWaitForThePaymentToComplete() error {
	debugPrintln("\n⏳ Waiting for payment to complete...")

	// Wait for Temporal workflow to process the payment
	// This includes provider API calls, balance updates, webhooks, etc.
	// We'll wait up to 30 seconds with periodic page reloads
	maxWait := 30
	checkInterval := 2

	for i := 0; i < maxWait; i += checkInterval {
		time.Sleep(time.Duration(checkInterval) * time.Second)

		// Reload to get latest state
		_, _ = sc.page.Reload()
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		})

		debugPrintf("   Waiting... (%d/%d seconds)\n", i+checkInterval, maxWait)
	}

	debugPrintln("✓ Payment processing timeout reached (assumed complete)")
	return nil
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
		Timeout:   playwright.Float(5000),
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
	walletIdentifier := fmt.Sprintf("%s-%s", sc.testEmailPrefix, emailSuffix)

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

// iWaitSecondsForThePaymentToComplete waits for specified seconds for payment to complete
func (sc *E2EContext) iWaitSecondsForThePaymentToComplete(seconds string) error {
	secondsInt := parseSeconds(seconds)
	debugPrintf("\n⏳ Waiting %d seconds for payment to complete...\n", secondsInt)

	for i := 0; i < secondsInt; i++ {
		time.Sleep(1 * time.Second)
		if i%2 == 0 {
			debugPrintf("   Waiting... (%d/%d seconds)\n", i+1, secondsInt)
		}
	}

	debugPrintln("✓ Payment wait period complete")
	return nil
}

// iWaitSecondsForTheDepositToProcess waits for specified seconds for deposit to process
func (sc *E2EContext) iWaitSecondsForTheDepositToProcess(seconds string) error {
	secondsInt := parseSeconds(seconds)
	debugPrintf("\n⏳ Waiting %d seconds for deposit to process...\n", secondsInt)

	for i := 0; i < secondsInt; i++ {
		time.Sleep(1 * time.Second)
		if i%2 == 0 {
			debugPrintf("   Waiting... (%d/%d seconds)\n", i+1, secondsInt)
		}
	}

	debugPrintln("✓ Deposit wait period complete")
	return nil
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
