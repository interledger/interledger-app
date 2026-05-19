package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
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

// iFilterTheWalletsListBy types searchTerm into the search input and submits
// the form, triggering a server-side search and page reload.
func (sc *E2EContext) iFilterTheWalletsListBy(searchTerm string) error {
	debugPrintf("🔍 Searching wallets by: %q\n", searchTerm)

	searchInput := sc.page.Locator("input[aria-label='Search wallets']")
	if err := searchInput.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("search input not found: %w", err)
	}

	if err := searchInput.Fill(searchTerm); err != nil {
		return fmt.Errorf("failed to fill search input: %w", err)
	}

	if err := searchInput.Press("Enter"); err != nil {
		return fmt.Errorf("failed to submit search: %w", err)
	}

	if err := sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("wallets page did not reload after search: %w", err)
	}

	debugPrintf("✓ Search applied: %q\n", searchTerm)
	return nil
}

// iFilterTheWalletsListByMyWalletName filters the wallets list using the
// current impersonated user's wallet name, which is searchable at the DB level.
func (sc *E2EContext) iFilterTheWalletsListByMyWalletName() error {
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

	debugPrintf("🔍 Searching wallets by wallet name %q (user: %s)\n", details.Name, email)
	return sc.iFilterTheWalletsListBy(details.Name)
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
// reads "1 result for …" (singular). This catches the regression where the filter
// is silently ignored and all wallets are returned instead of only the match.
func (sc *E2EContext) theWalletsListShouldShowExactlyOneResult() error {
	// The wallets page renders '<n> result(s) for "<search>"' only when a search
	// is active. After filtering to a single wallet the text must be singular.
	counter := sc.page.Locator(`p:has-text(" for ")`).First()
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
	if !strings.HasPrefix(trimmed, "1 result for") {
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
