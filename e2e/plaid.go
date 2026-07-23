package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Selectors for the mock Plaid Link overlay. The mockplaid window.Plaid shim
// (link-initialize.js) injects an overlay div containing an iframe whose content
// is served from cdn.plaid.com (link.html). The iframe body has native <select>
// controls plus Continue / Cancel buttons. We drive it through a FrameLocator
// because the iframe is cross-origin (cdn.plaid.com), not same-document.
const (
	plaidOverlaySelector  = "#mockplaid-overlay"
	plaidIframeSelector   = "#mockplaid-iframe"
	plaidBankSelector     = "#bank"
	plaidAccountSelector  = "#account"
	plaidContinueSelector = "#continue"
	plaidCancelSelector   = "#cancel"
)

// theMockplaidIsRunningAt records the mockplaid base URL and verifies both the
// mockplaid host and cdn.plaid.com (the Link iframe origin) resolve and that
// mockplaid is healthy. cdn.plaid.com is HSTS-preloaded, so it must be a trusted
// host redirect (make hosts/certs/trust), not a click-through bypass.
func (sc *E2EContext) theMockplaidIsRunningAt(urlStr string) error {
	sc.mockplaidBaseURL = strings.TrimSuffix(urlStr, "/")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse mockplaid URL: %w", err)
	}

	// Both the mockplaid host and the Link CDN host must resolve: the backend
	// SDK talks to mockplaid, the browser loads the Link iframe from cdn.plaid.com.
	if err := sc.ensureHostsResolve([]string{parsedURL.Hostname(), "cdn.plaid.com"}); err != nil {
		return err
	}

	debugPrintf("🔍 Verifying mockplaid health endpoint at %s...\n", urlStr)
	healthURL := strings.TrimSuffix(urlStr, "/") + "/health"
	return waitForHealthEndpoint(healthURL, 30*time.Second)
}

// iConnectViaPlaid drives the full overlay interaction: open /connect/bank (which
// auto-launches the Plaid flow), pick a bank + account in the iframe, and click
// Continue. It waits for the overlay to tear down (the shim removes it on both
// success and exit) but does NOT assert the outcome — callers assert via the
// snackbar and /accounts steps, because the outcome differs (success → redirect
// /accounts flash; already-linked → inline snackbar, no nav; failure → flash).
//
// bank is matched by its visible label (e.g. "Tartan Bank", "Platypus Bank");
// account is matched by its option value / key (e.g. "checking", "savings").
func (sc *E2EContext) iConnectViaPlaid(bank, account string) error {
	debugPrintf("\n🏦 Connecting via Plaid: bank=%q account=%q\n", bank, account)

	frame, err := sc.openPlaidOverlay(90 * time.Second)
	if err != nil {
		return err
	}
	if err := sc.selectPlaidBankAndContinue(frame, bank, account); err != nil {
		return err
	}
	// Wait only for the overlay to detach — do NOT wait for the /accounts redirect
	// or networkidle: the outcome differs and a long settle would dismiss the
	// inline "Account already linked" snackbar before a scenario can assert it.
	if err := sc.waitForPlaidOverlayGone(30 * time.Second); err != nil {
		return err
	}
	debugPrintf("   ✓ Plaid overlay closed; current URL: %s\n", sc.page.URL())
	return nil
}

// iCancelThePlaidOverlay clicks Cancel in the Plaid Link iframe and asserts the
// app navigates back to Home (the onExit(err=null) → onCancel path).
func (sc *E2EContext) iCancelThePlaidOverlay() error {
	debugPrintln("\n🚫 Cancelling the Plaid overlay...")

	frame, err := sc.openPlaidOverlay(90 * time.Second)
	if err != nil {
		return err
	}
	cancelBtn := frame.Locator(plaidCancelSelector)
	if err := cancelBtn.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("failed to click Plaid Cancel: %w", err)
	}

	if err := sc.waitForPlaidOverlayGone(15 * time.Second); err != nil {
		return err
	}

	// onCancel navigates to Home ('/'). Poll the URL until it lands there.
	base := strings.TrimSuffix(sc.baseURL, "/")
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if strings.TrimSuffix(sc.page.URL(), "/") == base {
			debugPrintf("   ✓ Cancelled; navigated back to Home: %s\n", sc.page.URL())
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("cancel did not navigate back to Home; current URL: %s", sc.page.URL())
}

// connectBankViaPlaid is a self-contained happy-path helper: link a US bank via
// the Plaid overlay and wait for the redirect to /accounts. Used both by the
// happy-path scenario and to repoint the legacy manual bank-connect step onto
// the Plaid flow (Plaid is the only US bank-link path in e2e).
func (sc *E2EContext) connectBankViaPlaid(bank, account string) error {
	frame, err := sc.openPlaidOverlay(90 * time.Second)
	if err != nil {
		return err
	}
	if err := sc.selectPlaidBankAndContinue(frame, bank, account); err != nil {
		return err
	}
	if err := sc.waitForPlaidOverlayGone(30 * time.Second); err != nil {
		return err
	}
	if err := sc.page.WaitForURL("**/accounts**", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(30000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("plaid-link-no-accounts-redirect")
		return fmt.Errorf("Plaid link did not redirect to /accounts: currentURL=%q: %w", sc.page.URL(), err)
	}
	debugPrintf("   ✓ Linked via Plaid, redirected to: %s\n", sc.page.URL())
	return nil
}

// iRemoveTheLinkedPlaidBankAccount opens the account detail page for the named
// account (from /accounts), clicks "Remove bank account", confirms in the
// dialog, and waits for the redirect back to /accounts. This soft-deletes the
// linked_account AND its plaid_links rows (DeleteLinkedAccount →
// SoftDeleteLinkByLinkedAccountID), freeing the (wallet, plaid_account_id)
// dedupe slot so the same Tartan account can be re-linked as a fresh link.
func (sc *E2EContext) iRemoveTheLinkedPlaidBankAccount(displayText string) error {
	debugPrintf("\n🗑️  Removing linked Plaid bank account %q...\n", displayText)

	if !strings.HasSuffix(strings.TrimSuffix(sc.page.URL(), "/"), "/accounts") {
		if _, err := sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(15000),
		}); err != nil {
			return fmt.Errorf("failed to navigate to /accounts: %w", err)
		}
	}

	accountLink := sc.page.Locator(fmt.Sprintf("a[href*='/accounts/']:has-text('%s')", displayText)).First()
	if err := accountLink.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(15000)}); err != nil {
		return fmt.Errorf("failed to open account detail for %q: %w", displayText, err)
	}

	removeBtn := sc.page.Locator("button:has-text('Remove bank account')")
	if err := removeBtn.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("failed to click 'Remove bank account': %w", err)
	}

	// Confirm in the dialog. The submit button is wired to the hidden
	// id='bank-delete' form (formName=delete).
	confirmBtn := sc.page.Locator("button[form='bank-delete']")
	if err := confirmBtn.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("failed to confirm bank account removal: %w", err)
	}

	if err := sc.page.WaitForURL("**/accounts", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(30000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("plaid-remove-no-redirect")
		return fmt.Errorf("removal did not redirect to /accounts: currentURL=%q: %w", sc.page.URL(), err)
	}
	debugPrintf("   ✓ Removed %q; back on /accounts\n", displayText)
	return nil
}

// --- internal helpers -----------------------------------------------------

// openPlaidOverlay navigates to /connect/bank and waits for the Link overlay to
// appear, returning a FrameLocator scoped to the iframe. It RETRIES navigation:
// right after KYC the wallet may not be activated yet (no US balance).
//
// Since WAL-535 the /connect/bank loader no longer redirects Home when the US
// balance is missing — it renders a "still being set up" card instead. We detect
// that card and keep retrying (fast) rather than burning the iframe-wait timeout
// on a page that can't auto-launch yet. We re-navigate until activation settles
// and the loader renders the auto-launching page, or until the deadline.
func (sc *E2EContext) openPlaidOverlay(timeout time.Duration) (playwright.FrameLocator, error) {
	deadline := time.Now().Add(timeout)
	iframe := sc.page.Locator(plaidIframeSelector)
	settingUp := sc.page.Locator("text=still being set up")
	for {
		if err := sc.gotoConnectBank(); err == nil && strings.Contains(sc.page.URL(), "/connect/bank") {
			// If the "still being set up" card is showing, the wallet isn't
			// activated yet — skip the iframe wait and retry after a short sleep.
			if visible, _ := settingUp.IsVisible(); !visible {
				// Activated (or still hydrating) — give the auto-launch a window.
				if werr := iframe.WaitFor(playwright.LocatorWaitForOptions{
					State:   playwright.WaitForSelectorStateVisible,
					Timeout: playwright.Float(15000),
				}); werr == nil {
					return sc.page.FrameLocator(plaidIframeSelector), nil
				}
			}
		}
		// Redirected away, still-setting-up, or the overlay never opened. Wait
		// for activation to progress, then retry.
		if time.Now().After(deadline) {
			_ = sc.iTakeAScreenshot("plaid-overlay-missing")
			return nil, fmt.Errorf(
				"Plaid Link overlay (%s) did not appear within %s (last URL %s); wallet likely not activated / no US balance yet",
				plaidIframeSelector, timeout, sc.page.URL())
		}
		time.Sleep(3 * time.Second)
	}
}

// gotoConnectBank navigates to /connect/bank, which auto-launches the Plaid flow.
func (sc *E2EContext) gotoConnectBank() error {
	target := sc.baseURL + "/connect/bank"
	if _, err := sc.page.Goto(target, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}); err != nil {
		return fmt.Errorf("failed to navigate to /connect/bank: %w", err)
	}
	return nil
}

// selectPlaidBankAndContinue picks the bank (by label) and account (by value)
// inside the already-open Link iframe and clicks Continue.
func (sc *E2EContext) selectPlaidBankAndContinue(frame playwright.FrameLocator, bank, account string) error {
	bankLabels := []string{bank}
	if _, err := frame.Locator(plaidBankSelector).SelectOption(playwright.SelectOptionValues{
		Labels: &bankLabels,
	}); err != nil {
		return fmt.Errorf("failed to select Plaid bank %q: %w", bank, err)
	}

	accountValues := []string{account}
	if _, err := frame.Locator(plaidAccountSelector).SelectOption(playwright.SelectOptionValues{
		Values: &accountValues,
	}); err != nil {
		return fmt.Errorf("failed to select Plaid account %q: %w", account, err)
	}

	_ = sc.iTakeAScreenshot("plaid-overlay-selection")

	if err := frame.Locator(plaidContinueSelector).Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("failed to click Plaid Continue: %w", err)
	}
	return nil
}

// waitForPlaidOverlayGone waits until the overlay div is removed from the DOM,
// which the shim does on both success and exit.
func (sc *E2EContext) waitForPlaidOverlayGone(timeout time.Duration) error {
	if err := sc.page.Locator(plaidOverlaySelector).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	}); err != nil {
		return fmt.Errorf("Plaid overlay did not close: %w", err)
	}
	return nil
}
