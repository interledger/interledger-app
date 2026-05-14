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
