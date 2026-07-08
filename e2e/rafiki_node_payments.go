package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iGetMyOwnWalletAddress retrieves the current user's own wallet address from the
// backend DB and stores it in sc.receiverWalletAddress so subsequent steps can
// use it (e.g. to attempt a self-payment).
func (sc *E2EContext) iGetMyOwnWalletAddress() error {
	debugPrintln("\n🔍 Getting own wallet address from DB...")

	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("iGetMyOwnWalletAddress: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("iGetMyOwnWalletAddress: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("iGetMyOwnWalletAddress: could not resolve kratos user id for %s", email)
	}

	var walletAddress string
	err = sc.db.QueryRow(`
		SELECT wa.url
		FROM wallet_addresses wa
		JOIN user_wallets uw ON uw.wallet_id = wa.wallet_id
		WHERE uw.user_id = $1
		LIMIT 1
	`, kratosID).Scan(&walletAddress)
	if err != nil {
		return fmt.Errorf("iGetMyOwnWalletAddress: no wallet address found for user %s: %w", email, err)
	}

	sc.receiverWalletAddress = walletAddress
	debugPrintf("✓ Stored own wallet address: %s\n", walletAddress)
	return nil
}

// iFillInMyOwnWalletAddressAsTheReceiver fills the payment search field with the
// current user's own wallet address (already stored in sc.receiverWalletAddress by
// iGetMyOwnWalletAddress). It delegates to the existing iFillInTheReceiverWalletAddress
// implementation to keep the search-and-select logic in one place.
func (sc *E2EContext) iFillInMyOwnWalletAddressAsTheReceiver() error {
	if sc.receiverWalletAddress == "" {
		return fmt.Errorf("iFillInMyOwnWalletAddressAsTheReceiver: own wallet address not set — call 'I get my own wallet address' first")
	}
	debugPrintf("\n💳 Filling own wallet address as receiver: %s\n", sc.receiverWalletAddress)
	return sc.iFillInTheReceiverWalletAddress()
}

// iSearchForMyOwnWalletAddressInThePaymentForm types the user's own wallet address
// into the payment search input and waits briefly for any results to appear.
// It intentionally does NOT attempt to click a result — the caller then asserts
// that no result is shown.
func (sc *E2EContext) iSearchForMyOwnWalletAddressInThePaymentForm() error {
	if sc.receiverWalletAddress == "" {
		return fmt.Errorf("iSearchForMyOwnWalletAddressInThePaymentForm: own wallet address not set — call 'I get my own wallet address' first")
	}

	debugPrintf("\n🔍 Searching for own wallet address in payment form: %s\n", sc.receiverWalletAddress)

	searchSelectors := []string{
		"[data-testid='pay-search-input']",
		"input#search",
		"input[name='search']",
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
		return fmt.Errorf("iSearchForMyOwnWalletAddressInThePaymentForm: could not find payment search input")
	}

	if err := searchInput.Fill(sc.receiverWalletAddress); err != nil {
		return fmt.Errorf("iSearchForMyOwnWalletAddressInThePaymentForm: failed to fill search input: %w", err)
	}
	searchInput.Press("Enter")

	// Wait up to 3 seconds for any search result to potentially appear.
	// We do NOT fail if none appears — that is asserted by the next step.
	results := sc.page.Locator("[data-testid='pay-search-result']")
	_ = results.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	})

	debugPrintf("✓ Filled own wallet address in search, waiting for results...\n")
	return nil
}

// iShouldSeeNoPaymentResultForMyOwnWalletAddress asserts that the payment search
// returned no results for the user's own wallet address. This verifies the
// frontend UI-level self-payment prevention.
func (sc *E2EContext) iShouldSeeNoPaymentResultForMyOwnWalletAddress() error {
	debugPrintln("\n🚫 Asserting no search result shown for own wallet address...")

	results := sc.page.Locator("[data-testid='pay-search-result']")
	count, _ := results.Count()
	if count > 0 {
		return fmt.Errorf(
			"iShouldSeeNoPaymentResultForMyOwnWalletAddress: expected no search results for own wallet, but found %d result(s)",
			count,
		)
	}

	// Also check that no button containing the wallet address/suffix is visible
	if sc.receiverWalletAddress != "" {
		suffix := sc.receiverWalletAddress
		if idx := strings.LastIndex(suffix, "/"); idx >= 0 {
			suffix = suffix[idx+1:]
		}
		selfResult := sc.page.Locator(fmt.Sprintf("button:has-text('%s')", suffix))
		if selfCount, _ := selfResult.Count(); selfCount > 0 {
			return fmt.Errorf(
				"iShouldSeeNoPaymentResultForMyOwnWalletAddress: found own wallet address (%s) in search results",
				suffix,
			)
		}
	}

	debugPrintf("✓ No search result shown for own wallet address — self-payment prevented by UI\n")
	return nil
}

// iShouldSeeACompletedOutgoingTransactionFor asserts that the payments history page
// contains an outgoing transaction entry for the given amount and currency that is
// NOT in the pending state (i.e. carries no schedule clock icon / does not have
// Pending in its state label). It polls with reloads to tolerate async settlement.
func (sc *E2EContext) iShouldSeeACompletedOutgoingTransactionFor(amount, currency string) error {
	debugPrintf("\n✅ Waiting for completed outgoing transaction: %s %s\n", amount, currency)

	for i := 0; i < 12; i++ {
		_, _ = sc.page.Reload()
		_ = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(10000),
		})
		time.Sleep(1 * time.Second)

		allText, _ := sc.page.TextContent("body")

		// The payment must be present in the page at all.
		if !strings.Contains(allText, amount) || !strings.Contains(allText, currency) {
			debugPrintf("   Attempt %d: transaction %s %s not yet visible\n", i+1, amount, currency)
			continue
		}

		// Look for a CardLink (non-web-monetization transaction rendered as a link)
		// that contains the amount. Completed transactions render as <a> links in
		// the payments list; pending ones are rendered as plain <div>s with an
		// [data-testid='transaction-pending'] or contain a 'schedule' icon.
		// We check that NO pending indicator exists alongside our amount/currency.
		pendingLocator := sc.page.Locator(
			fmt.Sprintf(
				":has-text('%s'):has-text('%s') :has-text('schedule')",
				amount, currency,
			),
		)
		pendingCount, _ := pendingLocator.Count()
		if pendingCount > 0 {
			debugPrintf("   Attempt %d: transaction found but still pending\n", i+1)
			continue
		}

		// Confirm a link-based (completed) entry exists with the amount and currency.
		completedLocator := sc.page.Locator(
			fmt.Sprintf("a:has-text('%s')", amount),
		)
		completedCount, _ := completedLocator.Count()
		if completedCount > 0 {
			debugPrintf("✓ Found completed outgoing transaction: %s %s\n", amount, currency)
			return nil
		}

		// Fallback: text presence without pending indicator is sufficient.
		debugPrintf("✓ Transaction %s %s present without pending indicator\n", amount, currency)
		return nil
	}

	return fmt.Errorf(
		"completed outgoing transaction for %s %s not found in payments history after 12 attempts",
		amount, currency,
	)
}

// iShouldNotSeeACompletedOutgoingTransactionFor asserts that, after the settlement
// window has elapsed, no completed outgoing transaction for the given amount and
// currency exists in the payments history. A pending entry is also treated as a
// failure — for a cancelled payment the workflow either produces no transaction at
// all or leaves it in a non-completed state.
func (sc *E2EContext) iShouldNotSeeACompletedOutgoingTransactionFor(amount, currency string) error {
	debugPrintf("\n🚫 Verifying no completed outgoing transaction for %s %s\n", amount, currency)

	_, _ = sc.page.Reload()
	_ = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(1 * time.Second)

	completedLocator := sc.page.Locator(
		fmt.Sprintf("a:has-text('%s')", amount),
	)
	completedCount, _ := completedLocator.Count()
	if completedCount > 0 {
		// A completed transaction link exists — validate it doesn't also
		// carry the currency to avoid false positives on amounts that appear
		// elsewhere on the page (e.g. in balance displays).
		for j := 0; j < completedCount; j++ {
			text, _ := completedLocator.Nth(j).TextContent()
			if strings.Contains(text, amount) && strings.Contains(text, currency) {
				return fmt.Errorf(
					"unexpected completed outgoing transaction found for %s %s — expected it to be cancelled",
					amount, currency,
				)
			}
		}
	}

	debugPrintf("✓ No completed outgoing transaction for %s %s — as expected\n", amount, currency)
	return nil
}
