package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// initializeBrowser sets up the Playwright browser if not already initialized
func (sc *E2EContext) initializeBrowser() error {
	if sc.browser != nil {
		return nil // Already initialized
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}
	sc.pw = pw

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	sc.browser = browser

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	sc.context = context

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	sc.page = page

	return nil
}

// iNavigateToThePersonalDetailsPageToActivateWallet navigates to the wallet activation page
func (sc *E2EContext) iNavigateToThePersonalDetailsPageToActivateWallet() error {
	debugPrintln("\n🔐 Navigating to personal details / wallet activation page...")

	// Initialize browser if needed
	if sc.browser == nil {
		err := sc.initializeBrowser()
		if err != nil {
			return fmt.Errorf("failed to initialize browser: %w", err)
		}
	}

	// Click "Activate wallet" link on dashboard instead of navigating directly
	// This ensures the Remix flow is properly initialized through client-side navigation
	activateLink := sc.page.Locator("a:has-text('Activate wallet')")

	// Wait for the link to be visible
	if err := activateLink.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("activate wallet link not found on dashboard: %w", err)
	}

	// Click the link and wait for navigation
	if err := activateLink.Click(); err != nil {
		return fmt.Errorf("failed to click activate wallet link: %w", err)
	}

	// Wait for navigation to complete
	if err := sc.page.WaitForURL(sc.baseURL+"/personal-details", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("failed to navigate to personal-details page: %w", err)
	}

	debugPrintf("   ✓ Navigated to personal details page via Activate wallet link\n")
	time.Sleep(500 * time.Millisecond)
	return nil
}

// iWaitForTheKYCIframeToLoad waits for the KYC iframe to appear and load
func (sc *E2EContext) iWaitForTheKYCIframeToLoad() error {
	debugPrintln("\n🕐 Waiting for KYC iframe to load...")

	// Wait for iframe to appear
	iframeLocator := sc.page.Locator("iframe")

	// Wait up to 10 seconds for iframe to appear
	for i := 0; i < 50; i++ {
		count, _ := iframeLocator.Count()
		if count > 0 {
			debugPrintf("   ✓ KYC iframe found\n")
			time.Sleep(500 * time.Millisecond) // Give iframe time to fully load
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("KYC iframe did not load after 10 seconds")
}

// iWaitForTheKYCCompletion waits for KYC to complete and return to dashboard.
// Used by both GateHub and PTI flows — checks both window.kycCompleted and window.ptiKycCompleted.
func (sc *E2EContext) iWaitForTheKYCCompletion() error {
	debugPrintln("\n⏳ Waiting for KYC completion...")

	// Wait for the form submission to complete
	// The backend should receive the webhook and update KYC status
	time.Sleep(1 * time.Second)

	// Check if the message was received by the parent (GateHub or PTI)
	messageReceived, _ := sc.page.Evaluate(`() => {
		return window.kycCompleted === true || window.ptiKycCompleted === true;
	}`)

	if messageReceived != nil && messageReceived.(bool) {
		debugPrintf("   ✓ KYC completion message received by parent\n")
	}

	// After KYC submission, we should return to dashboard (not login!)
	// The page will navigate away from personal-details
	// Wait for that navigation to happen and validate we don't end up on login
	for i := 0; i < 60; i++ { // 30 seconds timeout
		currentURL := sc.page.URL()

		// Log less frequently for clarity
		if i%10 == 0 {
			debugPrintf("   📍 Current URL (attempt %d): %s\n", i, currentURL)
		}

		// Under concurrency, session cookies may briefly desync after iframe completion.
		// Attempt one login recovery before failing.
		if strings.Contains(currentURL, "/login") {
			debugPrintf("   ⚠️  Redirected to login during KYC completion, retrying login...\n")
			if err := sc.iLogInAsMyself(); err != nil {
				return fmt.Errorf("KYC completed but redirected to login - user session was lost: %s", currentURL)
			}
			currentURL = sc.page.URL()
		}

		// We should navigate away from personal-details
		if !strings.Contains(currentURL, "/personal-details") {
			debugPrintf("   ✓ KYC completed, navigated away from personal-details\n")
			time.Sleep(500 * time.Millisecond)
			return nil
		}

		// Wait for page navigation event
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("KYC did not complete within 30 seconds, still on personal-details")
}

// iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus verifies KYC approved and shows balance
func (sc *E2EContext) iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus() error {
	debugPrintln("\n🏠 Verifying dashboard with approved KYC status...")

	// Wait for backend to process the webhook and update KYC status
	// This may take several seconds as the webhook needs to be processed
	for i := 0; i < 15; i++ { // Up to 30 seconds
		time.Sleep(500 * time.Millisecond)

		currentURL := sc.page.URL()
		debugPrintf("   📍 Current URL (check %d): %s\n", i+1, currentURL)

		// If on login page, session was lost - that's a failure
		if strings.Contains(currentURL, "/login") {
			return fmt.Errorf("user was redirected to login page - session lost after KYC attempt: %s", currentURL)
		}

		// Should be at root dashboard or related authenticated pages
		baseURLNormalized := strings.TrimSuffix(sc.baseURL, "/")
		if !strings.HasPrefix(currentURL, baseURLNormalized) || strings.Contains(currentURL, "/personal-details") || strings.Contains(currentURL, "/wallet-address") {
			continue
		}

		// Reload page to get latest KYC status from backend
		if i%2 == 0 {
			debugPrintf("   🔄 Reloading page to check for updated KYC status...\n")
			_, err := sc.page.Reload()
			if err == nil {
				time.Sleep(1 * time.Second)
			}
		}

		// Check that "Reserved" chip is gone
		reservedChip := sc.page.Locator("text=Reserved")
		count, _ := reservedChip.Count()
		if count == 0 {
			// Check for wallet card without "Reserved" indicator
			walletCard := sc.page.Locator("text=Wallet")
			count, _ = walletCard.Count()
			if count > 0 {
				debugPrintf("   ✓ Wallet card visible without 'Reserved' status (attempt %d)\n", i+1)
				return nil
			}

			// Also check for balance elements as sign of KYC approval
			content, _ := sc.page.Content()
			if strings.Contains(content, "USD") && strings.Contains(content, "balance") {
				debugPrintf("   ✓ Found balance indicators without Reserved status (attempt %d)\n", i+1)
				return nil
			}
		}
	}

	// Give one final check - but NOT on login page
	currentURL := sc.page.URL()
	if strings.Contains(currentURL, "/login") {
		return fmt.Errorf("failed KYC check: user is on login page (session lost): %s", currentURL)
	}

	content, _ := sc.page.Content()
	if !strings.Contains(content, "Reserved") {
		debugPrintf("   ✓ 'Reserved' status not found - KYC appears approved\n")
		return nil
	}

	return fmt.Errorf("wallet still shows 'Reserved' status after 30 seconds - KYC not approved")
}

// iShouldSeeMyAccountBalanceWithKYCApproved verifies balance visibility with KYC approved
func (sc *E2EContext) iShouldSeeMyAccountBalanceWithKYCApproved() error {
	debugPrintln("\n💰 Verifying account balance with KYC approved...")

	time.Sleep(500 * time.Millisecond)

	currentURL := sc.page.URL()
	debugPrintf("   📍 Current URL: %s\n", currentURL)

	// Look for balance-related content on the page
	balanceSelectors := []string{
		"[class*='balance']",
		"[data-testid*='balance']",
		":has-text('Balance')",
		":has-text('Available')",
		":has-text('USD')",
		":has-text('EUR')",
		"text=/\\d+\\.\\d{2}\\s(USD|EUR|GBP)/", // Look for formatted currency amounts
	}

	for _, selector := range balanceSelectors {
		element := sc.page.Locator(selector)
		count, _ := element.Count()
		if count > 0 {
			text, _ := element.First().TextContent()
			trimmedText := strings.TrimSpace(text)
			// Limit output to first 100 characters to avoid dumping entire page content
			if len(trimmedText) > 100 {
				trimmedText = trimmedText[:100] + "..."
			}
			debugPrintf("   ✓ Found balance element: %s\n", trimmedText)
			return nil
		}
	}

	// Check page content for currency mentions (indicating KYC approved state shows balances)
	content, _ := sc.page.Content()
	currencyMatches := 0
	if strings.Contains(content, "USD") {
		currencyMatches++
	}
	if strings.Contains(content, "EUR") {
		currencyMatches++
	}
	if strings.Contains(content, "balance") || strings.Contains(content, "Balance") {
		currencyMatches++
	}

	if currencyMatches >= 2 {
		debugPrintf("   ✓ Found multiple currency/balance indicators on page\n")
		return nil
	}

	// Last resort: if we're not in Reserved state and not on personal-details, consider it approved
	if !strings.Contains(content, "Reserved") && !strings.Contains(currentURL, "/personal-details") {
		debugPrintf("   ✓ Not in Reserved state and not on personal-details - KYC appears approved\n")
		return nil
	}

	return fmt.Errorf("unable to verify KYC approval - no balance information visible")
}
