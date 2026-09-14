package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func (sc *E2EContext) theAdminPortalIsRunningAt(urlStr string) error {
	sc.botanistBaseURL = strings.TrimSuffix(urlStr, "/")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse admin portal URL: %w", err)
	}

	if err := sc.ensureHostsResolve([]string{parsedURL.Hostname()}); err != nil {
		return err
	}

	healthURL := sc.botanistBaseURL + "/healthz"
	debugPrintf("🔍 Checking admin portal health at %s...\n", healthURL)
	if err := waitForHealthEndpoint(healthURL, 120*time.Second); err != nil {
		return fmt.Errorf("admin portal not ready: %w", err)
	}

	debugPrintf("✅ Admin portal is ready\n")
	return nil
}

func (sc *E2EContext) iNavigateToTheAdminPortal() error {
	if sc.browser == nil {
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
	}

	if sc.context != nil {
		if sc.page != nil {
			_ = sc.page.Close()
		}
		_ = sc.context.Close()
	}

	// Default viewport is 1280x720, which satisfies Tailwind's lg: breakpoint
	// so the desktop sidebar nav is rendered (not hidden).
	context, err := sc.browser.NewContext(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create browser context: %w", err)
	}
	sc.context = context

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	sc.page = page
	sc.page.OnConsole(func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" || msg.Type() == "warning" {
			debugPrintf("🔴 Browser %s: %s\n", msg.Type(), msg.Text())
		}
	})
	sc.page.OnPageError(func(err error) {
		debugPrintf("🔴 Browser page error: %v\n", err)
	})

	var navErr error
	for attempt := 0; attempt < 3; attempt++ {
		_, navErr = sc.page.Goto(sc.botanistBaseURL, playwright.PageGotoOptions{
			Timeout:   playwright.Float(30000),
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		})
		if navErr == nil {
			break
		}
		debugPrintf("   ⚠️  Navigation attempt %d failed: %v — retrying\n", attempt+1, navErr)
		time.Sleep(2 * time.Second)
	}
	if navErr != nil {
		return fmt.Errorf("failed to navigate to admin portal: %w", navErr)
	}

	if err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("admin portal did not reach network idle: %w", err)
	}

	debugPrintf("✓ Navigated to admin portal: %s\n", sc.botanistBaseURL)
	return nil
}

func (sc *E2EContext) theNavigationMenuShouldBeVisible() error {
	// The desktop sidebar nav contains NavLink elements rendered as <a> tags
	// wrapping <li> elements. At 1280px (lg breakpoint) the sidebar is always shown.
	// We check for the Home link as the anchor point for the whole nav.
	homeLink := sc.page.Locator("a[href='/']").First()
	if err := homeLink.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("nav-not-found")
		return fmt.Errorf("navigation menu not visible (home link not found): %w", err)
	}
	debugPrintln("✓ Navigation menu is visible")
	return nil
}

func (sc *E2EContext) theMenuItemShouldBeVisible(label string) error {
	// Each nav item is a NavLink wrapping an <li> with the label as text content.
	locator := sc.page.Locator(fmt.Sprintf("li:has-text('%s')", label)).First()
	if err := locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("menu item %q not visible: %w", label, err)
	}
	debugPrintf("✓ Menu item visible: %s\n", label)
	return nil
}

func (sc *E2EContext) thePageTitleShouldBe(expectedTitle string) error {
	title, err := sc.page.Title()
	if err != nil {
		return fmt.Errorf("failed to get page title: %w", err)
	}
	if title != expectedTitle {
		return fmt.Errorf("expected page title %q but got %q", expectedTitle, title)
	}
	debugPrintf("✓ Page title: %s\n", title)
	return nil
}

// iNavigateToTheBotanistWalletsPage navigates the current page to the /wallets route.
func (sc *E2EContext) iNavigateToTheBotanistWalletsPage() error {
	walletsURL := sc.botanistBaseURL + "/wallets"
	debugPrintf("\n🌐 Navigating to botanist wallets page: %s\n", walletsURL)

	_, err := sc.page.Goto(walletsURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to wallets page: %w", err)
	}

	if err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("wallets page did not reach network idle: %w", err)
	}

	debugPrintf("✓ On botanist wallets page\n")
	return nil
}

// fillWalletsFilterField types value into the named filter field on the
// wallets page. field must match one of wallets.tsx's FILTER_FIELDS names
// (e.g. "email", "walletAddress"), which are rendered as inputs with
// id="wallet-search-<field>".
func (sc *E2EContext) fillWalletsFilterField(field, value string) error {
	input := sc.page.Locator(fmt.Sprintf("#wallet-search-%s", field))
	if err := input.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("filter field %q not found: %w", field, err)
	}

	if err := input.Fill(value); err != nil {
		return fmt.Errorf("failed to fill filter field %q: %w", field, err)
	}

	return nil
}

// submitWalletsFilterForm submits the wallets filter form (all fields filled
// so far via fillWalletsFilterField) and waits for the server-side search to
// reload the page.
func (sc *E2EContext) submitWalletsFilterForm() error {
	submit := sc.page.Locator("button:has-text('Search')")
	if err := submit.Click(); err != nil {
		return fmt.Errorf("failed to submit wallets filter form: %w", err)
	}

	if err := sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("wallets page did not reload after search: %w", err)
	}

	return nil
}

// iFilterTheWalletsListByMyEmail filters the wallets list by the current
// impersonated user's email.
func (sc *E2EContext) iFilterTheWalletsListByMyEmail() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	debugPrintf("🔍 Filtering wallets by email %q\n", email)
	if err := sc.fillWalletsFilterField("email", email); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: email=%q\n", email)
	return nil
}

// iFilterTheWalletsListByMyWalletAddress filters the wallets list by the
// current impersonated user's wallet address.
func (sc *E2EContext) iFilterTheWalletsListByMyWalletAddress() error {
	details, err := sc.currentUserWalletDetails()
	if err != nil {
		return err
	}

	debugPrintf("🔍 Filtering wallets by wallet address %q\n", details.WalletAddress)
	if err := sc.fillWalletsFilterField("walletAddress", details.WalletAddress); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: walletAddress=%q\n", details.WalletAddress)
	return nil
}

// iFilterTheWalletsListByMyFirstName filters the wallets list by the current
// impersonated user's first name.
func (sc *E2EContext) iFilterTheWalletsListByMyFirstName() error {
	firstName := sc.kycFirstName()
	if firstName == "" {
		return fmt.Errorf("no first name set for current user")
	}

	debugPrintf("🔍 Filtering wallets by first name %q\n", firstName)
	if err := sc.fillWalletsFilterField("firstName", firstName); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: firstName=%q\n", firstName)
	return nil
}

// kycFirstName and kycLastName return the current user's first/last name as
// verified by the mock KYC provider — suffixed with the per-run test
// identifier (see xago_kyc.go) so that repeat runs of this scenario never
// collide on a shared static name (e.g. "Botanist"/"Tester" from the
// scenario's Background).
func (sc *E2EContext) kycFirstName() string {
	if sc.firstName == "" {
		return ""
	}
	return sc.firstName + sc.testIdentifier
}

func (sc *E2EContext) kycLastName() string {
	if sc.lastName == "" {
		return ""
	}
	return sc.lastName + sc.testIdentifier
}

// iFilterTheWalletsListByMyLastName filters the wallets list by the current
// impersonated user's last name.
func (sc *E2EContext) iFilterTheWalletsListByMyLastName() error {
	lastName := sc.kycLastName()
	if lastName == "" {
		return fmt.Errorf("no last name set for current user")
	}

	debugPrintf("🔍 Filtering wallets by last name %q\n", lastName)
	if err := sc.fillWalletsFilterField("lastName", lastName); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: lastName=%q\n", lastName)
	return nil
}

// iFilterTheWalletsListByMyPhoneNumber filters the wallets list by the current
// impersonated user's phone number.
func (sc *E2EContext) iFilterTheWalletsListByMyPhoneNumber() error {
	phone, err := sc.getCurrentUserPhone()
	if err != nil {
		return fmt.Errorf("cannot resolve current user phone: %w", err)
	}

	debugPrintf("🔍 Filtering wallets by phone number %q\n", phone)
	if err := sc.fillWalletsFilterField("phoneNumber", phone); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: phoneNumber=%q\n", phone)
	return nil
}

// iFilterTheWalletsListByMyProviderID filters the wallets list by the current
// impersonated user's linked-account provider ID.
func (sc *E2EContext) iFilterTheWalletsListByMyProviderID() error {
	details, err := sc.currentUserWalletDetails()
	if err != nil {
		return err
	}
	if details.ProviderID == "" {
		return fmt.Errorf("no provider ID found for current user's wallet")
	}

	debugPrintf("🔍 Filtering wallets by provider ID %q\n", details.ProviderID)
	if err := sc.fillWalletsFilterField("providerId", details.ProviderID); err != nil {
		return err
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}
	debugPrintf("✓ Filter applied: providerId=%q\n", details.ProviderID)
	return nil
}

// iFilterTheWalletsListByAllFilters fills every filter field at once before
// submitting, proving all six filters are ANDed together server-side rather
// than applied independently.
func (sc *E2EContext) iFilterTheWalletsListByAllFilters() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}
	phone, err := sc.getCurrentUserPhone()
	if err != nil {
		return fmt.Errorf("cannot resolve current user phone: %w", err)
	}
	firstName := sc.kycFirstName()
	lastName := sc.kycLastName()
	if firstName == "" || lastName == "" {
		return fmt.Errorf("no first/last name set for current user")
	}
	details, err := sc.currentUserWalletDetails()
	if err != nil {
		return err
	}
	if details.ProviderID == "" {
		return fmt.Errorf("no provider ID found for current user's wallet")
	}

	debugPrintf("🔍 Filtering wallets by all filters: firstName=%q lastName=%q email=%q phoneNumber=%q walletAddress=%q providerId=%q\n",
		firstName, lastName, email, phone, details.WalletAddress, details.ProviderID)

	fields := map[string]string{
		"firstName":     firstName,
		"lastName":      lastName,
		"email":         email,
		"phoneNumber":   phone,
		"walletAddress": details.WalletAddress,
		"providerId":    details.ProviderID,
	}
	for _, field := range []string{"firstName", "lastName", "email", "phoneNumber", "walletAddress", "providerId"} {
		if err := sc.fillWalletsFilterField(field, fields[field]); err != nil {
			return err
		}
	}
	if err := sc.submitWalletsFilterForm(); err != nil {
		return err
	}

	debugPrintf("✓ All filters applied\n")
	return nil
}

// currentUserWalletDetails resolves the current impersonated user's wallet
// details (ID, name, wallet address) via Kratos + the backend DB.
func (sc *E2EContext) currentUserWalletDetails() (*WalletDetails, error) {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve current user email: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return nil, fmt.Errorf("cannot resolve kratos ID for email %q", email)
	}

	details, err := sc.getWalletDetailsForUser(kratosID)
	if err != nil {
		return nil, fmt.Errorf("cannot get wallet details for user %q: %w", email, err)
	}

	return details, nil
}

// myWalletShouldAppearInTheWalletsList asserts that a table row containing the
// current user's email is visible in the wallets table.
func (sc *E2EContext) myWalletShouldAppearInTheWalletsList() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot check wallet visibility: %w", err)
	}

	debugPrintf("🔎 Looking for wallet with email %q in table\n", email)

	// Poll briefly — after signup there may be a short propagation delay.
	locator := sc.page.Locator(fmt.Sprintf("td:has-text('%s')", email)).First()
	if err := locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(15000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("wallet-not-found")
		return fmt.Errorf("wallet with email %q not visible in wallets table: %w", email, err)
	}

	debugPrintf("✓ Wallet for %q is visible in the list\n", email)
	return nil
}

// theWalletsListShouldShowExactlyOneResult asserts that the search result counter
// reads "1 result" (singular). This catches the regression where the filter
// is silently ignored and all wallets are returned instead of only the match.
func (sc *E2EContext) theWalletsListShouldShowExactlyOneResult() error {
	// The wallets page renders '<n> result(s)' only when a filter is active
	// (see wallets.tsx's hasFilter block). After filtering to a single wallet
	// the text must read exactly "1 result".
	counter := sc.page.Locator(`p:text-matches("^\\d+ results?$")`).First()
	if err := counter.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("result-counter-not-found")
		return fmt.Errorf("result counter not visible after search: %w", err)
	}

	text, err := counter.TextContent()
	if err != nil {
		return fmt.Errorf("failed to read result counter: %w", err)
	}

	trimmed := strings.TrimSpace(text)
	if trimmed != "1 result" {
		_ = sc.iTakeAScreenshot("unexpected-result-count")
		return fmt.Errorf("filter did not narrow to 1 result — counter says: %q", trimmed)
	}

	debugPrintf("✓ Result counter shows exactly 1 result: %q\n", trimmed)
	return nil
}

// theWalletsListShouldHaveMoreThanOneResult asserts that the unfiltered wallets
// table contains at least two rows. This is a pre-condition for the filter tests:
// if the list already has only one wallet, filtering to one result would not prove
// the filter is working.
func (sc *E2EContext) theWalletsListShouldHaveMoreThanOneResult() error {
	// Data rows are <tr> elements inside <tbody> that are NOT the pagination row.
	rows := sc.page.Locator(`tbody tr:not([aria-label="Pagination"])`)
	count, err := rows.Count()
	if err != nil {
		return fmt.Errorf("failed to count wallet rows: %w", err)
	}
	if count < 2 {
		_ = sc.iTakeAScreenshot("too-few-wallets-unfiltered")
		return fmt.Errorf("expected at least 2 wallets in the unfiltered list, got %d — filter test would be meaningless", count)
	}

	debugPrintf("✓ Unfiltered list has %d wallet rows\n", count)
	return nil
}

func (sc *E2EContext) iNavigateToMyWalletProfileInAdminPortal() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("cannot resolve kratos ID for email %q", email)
	}

	details, err := sc.getWalletDetailsForUser(kratosID)
	if err != nil {
		return fmt.Errorf("cannot get wallet details for user %q: %w", email, err)
	}

	profileURL := sc.botanistBaseURL + "/wallet/" + details.ID + "/profile"
	debugPrintf("\n🌐 Navigating to botanist wallet profile page: %s\n", profileURL)

	_, err = sc.page.Goto(profileURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to wallet profile page: %w", err)
	}

	if err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("wallet profile page did not reach network idle: %w", err)
	}

	// Wait for React to finish client-side hydration before interacting.
	// The root layout sets data-hydrated="true" on <html> inside a useEffect,
	// which only runs after hydration is complete.
	hydrationLocator := sc.page.Locator("html[data-hydrated='true']")
	if err = hydrationLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("React hydration did not complete on wallet profile page: %w", err)
	}

	debugPrintf("✓ On botanist wallet profile page\n")
	return nil
}

func (sc *E2EContext) theResetAuthenticatorButtonShouldBeVisible() error {
	btn := sc.page.Locator("button:has-text('Reset authenticator')").First()
	if err := btn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-authenticator-not-visible")
		return fmt.Errorf("reset authenticator button not visible: %w", err)
	}

	debugPrintln("✓ Reset authenticator button is visible")
	return nil
}

func (sc *E2EContext) theResetAuthenticatorButtonShouldNotBeVisible() error {
	btn := sc.page.Locator("button:has-text('Reset authenticator')").First()

	count, err := btn.Count()
	if err != nil {
		return fmt.Errorf("failed to inspect reset authenticator button: %w", err)
	}

	if count == 0 {
		debugPrintln("✓ Reset authenticator button is hidden")
		return nil
	}

	isVisible, err := btn.IsVisible()
	if err != nil {
		return fmt.Errorf("failed to check reset authenticator visibility: %w", err)
	}

	if isVisible {
		_ = sc.iTakeAScreenshot("reset-authenticator-still-visible")
		return fmt.Errorf("reset authenticator button should be hidden after reset")
	}

	debugPrintln("✓ Reset authenticator button is hidden")
	return nil
}

func (sc *E2EContext) iClickTheResetAuthenticatorButton() error {
	btn := sc.page.Locator("button:has-text('Reset authenticator')").First()
	if err := btn.Click(); err != nil {
		return fmt.Errorf("failed to click reset authenticator button: %w", err)
	}

	debugPrintln("✓ Clicked reset authenticator button")
	return nil
}

func (sc *E2EContext) theAuthenticatorResetConfirmationModalShouldBeVisible() error {
	modalTitle := sc.page.Locator("text=Reset authenticator for this wallet owner?").First()
	if err := modalTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-authenticator-modal-not-visible")
		return fmt.Errorf("confirmation modal not visible: %w", err)
	}

	debugPrintln("✓ Reset authenticator confirmation modal is visible")
	return nil
}

func (sc *E2EContext) iConfirmTheAuthenticatorReset() error {
	confirmButton := sc.page.Locator("button:has-text('Confirm reset')").First()
	if err := confirmButton.Click(); err != nil {
		return fmt.Errorf("failed to click confirm reset button: %w", err)
	}

	modalTitle := sc.page.Locator("text=Reset authenticator for this wallet owner?").First()
	if err := modalTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(20000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-authenticator-modal-stuck")
		return fmt.Errorf("confirmation modal did not close after reset: %w", err)
	}

	debugPrintln("✓ Confirmed authenticator reset")
	return nil
}

func (sc *E2EContext) myTotpShouldBeDisabled() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	_, err = sc.getTOTPSecretForEmail(email)
	if err == nil {
		return fmt.Errorf("expected TOTP to be disabled for %q but secret is still present", email)
	}

	debugPrintf("✓ TOTP credentials are removed for %q\n", email)
	return nil
}

func (sc *E2EContext) anAuthenticatorResetAuditLogEntryShouldExist() error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("audit assertion: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("cannot resolve kratos ID for email %q", email)
	}

	details, err := sc.getWalletDetailsForUser(kratosID)
	if err != nil {
		return fmt.Errorf("cannot get wallet details for user %q: %w", email, err)
	}

	var count int
	err = sc.db.QueryRow(`
		SELECT COUNT(*)
		FROM admin_audit_log
		WHERE wallet_id = $1
		AND operation LIKE '%Delete2FATotpEnrollment'
	`, details.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query admin audit log: %w", err)
	}

	if count < 1 {
		return fmt.Errorf("expected at least one authenticator reset audit log entry for wallet %s", details.ID)
	}

	debugPrintf("✓ Found %d authenticator reset audit log entry(ies) for wallet %s\n", count, details.ID)
	return nil
}

func (sc *E2EContext) theResetSmsOtpButtonShouldBeVisible() error {
	btn := sc.page.Locator("button:has-text('Reset SMS OTP')").First()
	if err := btn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-sms-otp-not-visible")
		return fmt.Errorf("reset SMS OTP button not visible: %w", err)
	}

	debugPrintln("✓ Reset SMS OTP button is visible")
	return nil
}

func (sc *E2EContext) theResetSmsOtpButtonShouldNotBeVisible() error {
	btn := sc.page.Locator("button:has-text('Reset SMS OTP')").First()

	count, err := btn.Count()
	if err != nil {
		return fmt.Errorf("failed to inspect reset SMS OTP button: %w", err)
	}

	if count == 0 {
		debugPrintln("✓ Reset SMS OTP button is hidden")
		return nil
	}

	isVisible, err := btn.IsVisible()
	if err != nil {
		return fmt.Errorf("failed to check reset SMS OTP visibility: %w", err)
	}

	if isVisible {
		_ = sc.iTakeAScreenshot("reset-sms-otp-still-visible")
		return fmt.Errorf("reset SMS OTP button should be hidden after reset")
	}

	debugPrintln("✓ Reset SMS OTP button is hidden")
	return nil
}

func (sc *E2EContext) iClickTheResetSmsOtpButton() error {
	btn := sc.page.Locator("button:has-text('Reset SMS OTP')").First()
	if err := btn.Click(); err != nil {
		return fmt.Errorf("failed to click reset SMS OTP button: %w", err)
	}

	debugPrintln("✓ Clicked reset SMS OTP button")
	return nil
}

const smsOtpResetModalTitle = "text=Reset SMS OTP verification for this wallet owner?"

func (sc *E2EContext) theSmsOtpResetConfirmationModalShouldBeVisible() error {
	modalTitle := sc.page.Locator(smsOtpResetModalTitle).First()
	if err := modalTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-sms-otp-modal-not-visible")
		return fmt.Errorf("confirmation modal not visible: %w", err)
	}

	debugPrintln("✓ Reset SMS OTP confirmation modal is visible")
	return nil
}

func (sc *E2EContext) iConfirmTheSmsOtpReset() error {
	confirmButton := sc.page.Locator("button:has-text('Confirm reset')").First()
	if err := confirmButton.Click(); err != nil {
		return fmt.Errorf("failed to click confirm reset button: %w", err)
	}

	modalTitle := sc.page.Locator(smsOtpResetModalTitle).First()
	if err := modalTitle.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(20000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("reset-sms-otp-modal-stuck")
		return fmt.Errorf("confirmation modal did not close after reset: %w", err)
	}

	debugPrintln("✓ Confirmed SMS OTP reset")
	return nil
}

func (sc *E2EContext) mySmsOtpShouldBeNotVerified() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	var lastErr error
	for {
		verified, err := sc.kratosPhoneIsVerified(email)
		switch {
		case err != nil:
			lastErr = err
		case !verified:
			debugPrintf("✓ SMS OTP is not verified for %q\n", email)
			return nil
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return fmt.Errorf("could not read phoneVerified for %q: %w", email, lastErr)
			}
			return fmt.Errorf("expected SMS OTP to be unverified for %q but phoneVerified is still true", email)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (sc *E2EContext) anSmsOtpResetAuditLogEntryShouldExist() error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("audit assertion: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("cannot resolve current user email: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("cannot resolve kratos ID for email %q", email)
	}

	details, err := sc.getWalletDetailsForUser(kratosID)
	if err != nil {
		return fmt.Errorf("cannot get wallet details for user %q: %w", email, err)
	}

	var count int
	err = sc.db.QueryRow(`
		SELECT COUNT(*)
		FROM admin_audit_log
		WHERE wallet_id = $1
		AND operation LIKE '%ResetUserPhoneVerification'
	`, details.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query admin audit log: %w", err)
	}

	if count < 1 {
		return fmt.Errorf("expected at least one SMS OTP reset audit log entry for wallet %s", details.ID)
	}

	debugPrintf("✓ Found %d SMS OTP reset audit log entry(ies) for wallet %s\n", count, details.ID)
	return nil
}
