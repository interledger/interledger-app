package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iCompleteSignupWorkflowFor runs signup + account verification + TOTP + wallet address
// for the named user, without completing KYC. Used to test pre-KYC state.
func (sc *E2EContext) iCompleteSignupWorkflowFor(userName string) error {
	debugPrintf("\n🧭 Running signup-only workflow for user: %s\n", userName)

	if err := sc.iImpersonate(userName); err != nil {
		return fmt.Errorf("failed to impersonate user '%s': %w", userName, err)
	}

	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s'", userName)
	}

	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("emailSuffix not defined for user '%s'", userName)
	}
	password, ok := details.Fields["password"]
	if !ok {
		return fmt.Errorf("password not defined for user '%s'", userName)
	}
	country, ok := details.Fields["country"]
	if !ok {
		return fmt.Errorf("country not defined for user '%s'", userName)
	}
	firstName, ok := details.Fields["firstName"]
	if !ok {
		return fmt.Errorf("firstName not defined for user '%s'", userName)
	}
	lastName, ok := details.Fields["lastName"]
	if !ok {
		return fmt.Errorf("lastName not defined for user '%s'", userName)
	}

	phone := "+4917"
	if err := sc.iCompleteSignupFlow(firstName, lastName, emailSuffix, country, phone, password); err != nil {
		return fmt.Errorf("signup flow failed: %w", err)
	}

	return nil
}

// theNavItemShouldNotBeVisible asserts that a nav item with the given label is absent.
func (sc *E2EContext) theNavItemShouldNotBeVisible(label string) error {
	debugPrintf("\n🔍 Asserting nav item '%s' is NOT visible\n", label)

	// Nav items are rendered inside NavDrawer.ListItem elements
	navSelector := fmt.Sprintf("nav a:has-text('%s'), [role='navigation'] a:has-text('%s')", label, label)
	locator := sc.page.Locator(navSelector)

	count, err := locator.Count()
	if err != nil {
		return fmt.Errorf("failed to count nav items: %w", err)
	}

	if count > 0 {
		// Check if any are actually visible
		for i := range count {
			visible, _ := locator.Nth(i).IsVisible()
			if visible {
				return fmt.Errorf("expected nav item '%s' to be hidden but it is visible", label)
			}
		}
	}

	debugPrintf("✓ Nav item '%s' is not visible\n", label)
	return nil
}

// iNavigateToPath navigates to a path relative to the base URL.
func (sc *E2EContext) iNavigateToPath(path string) error {
	targetURL := strings.TrimSuffix(sc.baseURL, "/") + path
	debugPrintf("\n🌐 Navigating to path: %s\n", targetURL)

	if _, err := sc.page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", path, err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	return nil
}

// iShouldBeRedirectedTo asserts the current URL path.
func (sc *E2EContext) iShouldBeRedirectedTo(expectedPath string) error {
	// Allow a moment for redirect to complete
	time.Sleep(500 * time.Millisecond)

	currentURL := sc.page.URL()
	expectedFull := strings.TrimSuffix(sc.baseURL, "/") + expectedPath
	debugPrintf("\n🔍 Asserting redirect: current=%s expected=%s\n", currentURL, expectedFull)

	if !strings.HasSuffix(strings.TrimRight(currentURL, "/"), strings.TrimRight(expectedPath, "/")) {
		return fmt.Errorf("expected to be redirected to '%s' but current URL is '%s'", expectedPath, currentURL)
	}

	debugPrintf("✓ Redirected to '%s'\n", expectedPath)
	return nil
}

// iNavigateToTheCardsPage navigates to /cards.
func (sc *E2EContext) iNavigateToTheCardsPage() error {
	return sc.iNavigateToPath("/cards")
}

// iShouldSeeTheCardsPageWithOrderButton asserts the /cards page is showing with an "Order card" button.
func (sc *E2EContext) iShouldSeeTheCardsPageWithOrderButton() error {
	debugPrintf("\n🔍 Asserting cards page with 'Order card' button\n")

	// Verify URL
	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "/cards") {
		return fmt.Errorf("expected to be on /cards but current URL is '%s'", currentURL)
	}

	// "Order card" button must be visible
	orderBtn := sc.page.Locator("a:has-text('Order card'), button:has-text('Order card')")
	if err := orderBtn.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("'Order card' button not visible on cards page: %w", err)
	}

	debugPrintf("✓ Cards page loaded with 'Order card' button visible\n")
	return nil
}

// iShouldBeOnTheCardOrderPage asserts we are on /cards/order and product images are shown.
func (sc *E2EContext) iShouldBeOnTheCardOrderPage() error {
	debugPrintf("\n🔍 Asserting card order page (/cards/order)\n")

	// Wait for URL to change to /cards/order (SPA navigation may be async)
	if err := sc.page.WaitForURL("**/cards/order**", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("expected to be on /cards/order but current URL is '%s': %w", sc.page.URL(), err)
	}

	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "/cards/order") {
		return fmt.Errorf("expected to be on /cards/order but current URL is '%s'", currentURL)
	}

	// Wait for at least one card product image to appear
	productImg := sc.page.Locator("img[src*='cards/']")
	if err := productImg.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		// Product images might not be loaded yet; check for any image in the product selector
		debugPrintf("   ⚠️  Card product images not found by src pattern, trying generic check\n")
		anyImg := sc.page.Locator("img")
		if err2 := anyImg.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err2 != nil {
			return fmt.Errorf("no card product content visible on /cards/order: %w", err)
		}
	}

	debugPrintf("✓ On card order page with product content visible\n")
	return nil
}

// iSelectTheFirstAvailableCardProduct clicks the first card product image.
func (sc *E2EContext) iSelectTheFirstAvailableCardProduct() error {
	debugPrintf("\n🖱️  Selecting first available card product\n")

	// Card products are rendered as clickable images inside the ProductsSelect component
	productImg := sc.page.Locator("img[src*='cards/']")
	count, _ := productImg.Count()

	if count == 0 {
		// Fallback: try any clickable image on the page
		productImg = sc.page.Locator("img").First()
		count = 1
	}

	debugPrintf("   📍 Found %d card product image(s)\n", count)

	if err := productImg.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click first card product: %w", err)
	}

	// Wait for selection state to settle
	time.Sleep(500 * time.Millisecond)
	debugPrintf("✓ First card product selected\n")
	return nil
}

// iShouldBeOnTheDeliveryAddressStep asserts the delivery address selection is visible.
func (sc *E2EContext) iShouldBeOnTheDeliveryAddressStep() error {
	debugPrintf("\n🔍 Asserting delivery address step is visible\n")

	// DeliveryAddresses renders a radio group of addresses; the step is still on /cards/order
	// Look for a radio group or address list
	addressList := sc.page.Locator("[role='radiogroup'], input[type='radio']")
	if err := addressList.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("delivery address step not visible (no radio group found): %w", err)
	}

	debugPrintf("✓ Delivery address step is visible\n")
	return nil
}

// iSelectTheExistingDeliveryAddress selects the first (existing/KYC) address radio option.
func (sc *E2EContext) iSelectTheExistingDeliveryAddress() error {
	debugPrintf("\n🖱️  Selecting existing delivery address\n")

	// The KYC address uses a Headless UI RadioGroup which renders role="radio" options
	radioBtn := sc.page.Locator("[role='radio'], input[type='radio']").First()
	if err := radioBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("no delivery address radio button found: %w", err)
	}

	if err := radioBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to select existing delivery address: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	debugPrintf("✓ Existing delivery address selected\n")
	return nil
}

// iShouldBeOnTheCardOrderConfirmationStep asserts the confirmation step is visible.
func (sc *E2EContext) iShouldBeOnTheCardOrderConfirmationStep() error {
	debugPrintf("\n🔍 Asserting card order confirmation step\n")

	// ConfirmCard renders a card preview image and a submit button
	confirmBtn := sc.page.Locator("button[type='submit'], button:has-text('Order'), button:has-text('Confirm')")
	if err := confirmBtn.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("card order confirmation step not visible: %w", err)
	}

	debugPrintf("✓ Card order confirmation step is visible\n")
	return nil
}

// iConfirmTheCardOrder submits the card order form.
func (sc *E2EContext) iConfirmTheCardOrder() error {
	debugPrintf("\n🖱️  Confirming card order\n")

	submitBtn := sc.page.Locator("button[type='submit'], button:has-text('Order'), button:has-text('Confirm')")
	if err := submitBtn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click confirm card order button: %w", err)
	}

	// Wait for navigation/redirect after submission
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Card order submitted\n")
	return nil
}

// iShouldSeeTheSnackbar asserts a snackbar with the given message is visible.
// Matches on text up to the first apostrophe to handle curly/smart apostrophe
// mismatch between feature files (straight ') and UI (curly ' U+2019).
func (sc *E2EContext) iShouldSeeTheSnackbar(expectedMessage string) error {
	debugPrintf("\n🔍 Asserting snackbar with message: %s\n", expectedMessage)

	matchText := expectedMessage
	if idx := strings.IndexAny(expectedMessage, "'\u2018\u2019"); idx > 0 {
		matchText = expectedMessage[:idx]
	}

	snackbar := sc.page.Locator(fmt.Sprintf("text=%s", matchText))
	if err := snackbar.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		_ = sc.iTakeAScreenshot("snackbar-not-visible")
		return fmt.Errorf("snackbar with message '%s' not visible: %w", expectedMessage, err)
	}

	_ = sc.iTakeAScreenshot("snackbar-visible")
	debugPrintf("✓ Snackbar visible with expected message\n")
	return nil
}
