package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	requiredFields := []string{"emailSuffix", "password", "country", "firstName", "lastName"}
	for _, f := range requiredFields {
		if _, ok := details.Fields[f]; !ok {
			return fmt.Errorf("%s not defined for user '%s'", f, userName)
		}
	}

	phone := "+4917"
	if err := sc.iCompleteSignupFlow(details.Fields["firstName"], details.Fields["lastName"], details.Fields["emailSuffix"], details.Fields["country"], phone, details.Fields["password"]); err != nil {
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
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})

	return nil
}

// iShouldBeRedirectedTo asserts the current URL path.
// Uses a short timeout — by the time this step runs after iShouldSeeTheSnackbar,
// the redirect has already happened. A long WaitForURL here would consume the
// snackbar's 4s auto-dismiss window before the snackbar assertion runs.
func (sc *E2EContext) iShouldBeRedirectedTo(expectedPath string) error {
	debugPrintf("\n🔍 Asserting redirect to: %s\n", expectedPath)

	pattern := "**" + strings.TrimRight(expectedPath, "/")
	if err := sc.page.WaitForURL(pattern, playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(3000),
	}); err != nil {
		return fmt.Errorf("expected to be redirected to '%s' but current URL is '%s': %w", expectedPath, sc.page.URL(), err)
	}

	debugPrintf("✓ Redirected to '%s'\n", expectedPath)
	return nil
}

// iNavigateToTheCardsPage navigates to /cards.
func (sc *E2EContext) iNavigateToTheCardsPage() error {
	return sc.iNavigateToPath("/cards")
}

// iShouldSeeTheCardsPageWithButton asserts the /cards page is showing with the given button.
func (sc *E2EContext) iShouldSeeTheCardsPageWithButton(buttonText string) error {
	debugPrintf("\n🔍 Asserting cards page with '%s' button\n", buttonText)

	// Verify URL
	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "/cards") {
		return fmt.Errorf("expected to be on /cards but current URL is '%s'", currentURL)
	}

	// Button must be visible (could be a link or button element)
	btn := sc.page.Locator(fmt.Sprintf("a:has-text('%s'), button:has-text('%s')", buttonText, buttonText))
	if err := btn.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("'%s' button not visible on cards page: %w", buttonText, err)
	}

	debugPrintf("✓ Cards page loaded with '%s' button visible\n", buttonText)
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
		return fmt.Errorf("no card product images visible on /cards/order: %w", err)
	}

	debugPrintf("✓ On card order page with product content visible\n")
	return nil
}

// iSelectTheFirstAvailableCardProduct clicks the first card product image.
func (sc *E2EContext) iSelectTheFirstAvailableCardProduct() error {
	debugPrintf("\n🖱️  Selecting first available card product\n")

	// Card products are rendered as clickable images inside the ProductsSelect component
	productImg := sc.page.Locator("img[src*='cards/']")
	if err := productImg.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("no card product images found to select: %w", err)
	}

	count, _ := productImg.Count()
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

	// ConfirmCard renders a "Confirm Card" heading, a card preview, and a submit button bound to form#confirm-card
	confirmBtn := sc.page.Locator("button[form='confirm-card'], :has-text('Confirm Card') >> button[type='submit']")
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

	submitBtn := sc.page.Locator("button[form='confirm-card']")
	if err := submitBtn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click confirm card order button: %w", err)
	}

	// Do NOT wait for networkidle here — the snackbar auto-dismisses in 4s and
	// waiting too long would cause the snackbar assertion to miss it.
	// iShouldBeRedirectedTo handles the navigation wait via WaitForURL.
	debugPrintf("✓ Card order submitted\n")
	return nil
}

// getActiveCardID queries MockGatehub for the first active card belonging to gatehubUserID.
func (sc *E2EContext) getActiveCardID(gatehubUserID string) (string, error) {
	baseURL := mockgatehubCardTxBaseURL
	if sc.mockgatehubBaseURL != "" {
		baseURL = sc.mockgatehubBaseURL
	}

	reqURL := fmt.Sprintf("%s/ui/actions/card-transaction/cards?userID=%s", baseURL, url.QueryEscape(gatehubUserID))
	req, err := http.NewRequestWithContext(context.Background(), "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("getActiveCardID: build request: %w", err)
	}

	resp, err := mockgatehubHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("getActiveCardID: request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("getActiveCardID: read body: %w", err)
	}

	var cards []struct {
		ID string `json:"id"`
	}
	if err = json.Unmarshal(body, &cards); err != nil {
		return "", fmt.Errorf("getActiveCardID: decode cards: %w (body: %s)", err, string(body))
	}

	if len(cards) == 0 {
		return "", fmt.Errorf("getActiveCardID: no active cards for user %s", gatehubUserID)
	}

	return cards[0].ID, nil
}

// waitForCardInMockgatehub polls MockGatehub until an active card exists for gatehubUserID.
func (sc *E2EContext) waitForCardInMockgatehub(gatehubUserID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		cardID, err := sc.getActiveCardID(gatehubUserID)
		if err == nil && cardID != "" {
			debugPrintf("   ✓ Card %s is active in MockGatehub for user %s\n", cardID, gatehubUserID)
			return nil
		}
		debugPrintf("   ⏳ Waiting for card in MockGatehub for user %s...\n", gatehubUserID)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("waitForCardInMockgatehub: timed out waiting for card for user %s", gatehubUserID)
}

// iOrderAPhysicalCardFor is a composite step that navigates through the full card
// order flow for the named user and waits until the card appears in MockGatehub.
func (sc *E2EContext) iOrderAPhysicalCardFor(userName string) error {
	debugPrintf("\n🃏 Ordering physical card for user: %s\n", userName)

	if err := sc.iImpersonate(userName); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: impersonate: %w", err)
	}
	if err := sc.iNavigateToTheCardsPage(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: navigate: %w", err)
	}
	if err := sc.iClickTheButton("Order card"); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: click order card: %w", err)
	}
	if err := sc.iShouldBeOnTheCardOrderPage(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: card order page: %w", err)
	}
	if err := sc.iSelectTheFirstAvailableCardProduct(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: select product: %w", err)
	}
	if err := sc.iClickTheButton("Physical"); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: click physical: %w", err)
	}
	if err := sc.iShouldBeOnTheDeliveryAddressStep(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: delivery address step: %w", err)
	}
	if err := sc.iSelectTheExistingDeliveryAddress(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: select address: %w", err)
	}
	if err := sc.iClickTheButton("Confirm"); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: click confirm: %w", err)
	}
	if err := sc.iShouldBeOnTheCardOrderConfirmationStep(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: confirmation step: %w", err)
	}
	if err := sc.iConfirmTheCardOrder(); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: confirm order: %w", err)
	}

	email, err := sc.getEmailForUser(userName)
	if err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: get email: %w", err)
	}
	gatehubUserID, err := sc.getGatehubUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: get gatehub user ID: %w", err)
	}
	if err = sc.waitForCardInMockgatehub(gatehubUserID, 60*time.Second); err != nil {
		return fmt.Errorf("iOrderAPhysicalCardFor: wait for card: %w", err)
	}

	debugPrintf("✓ Virtual card ordered and confirmed in MockGatehub for user: %s\n", userName)
	return nil
}

// iShouldSeeTheSnackbar asserts a snackbar with the given message is visible.
// Matches on text up to the first apostrophe to handle curly/smart apostrophe
// mismatch between feature files (straight ') and UI (curly ' U+2019).
func (sc *E2EContext) iShouldSeeTheSnackbar(expectedMessage string) error {
	debugPrintf("\n🔍 Asserting snackbar with message: %s\n", expectedMessage)

	// Normalize curly apostrophes to straight ones, then take the substring before
	// the first apostrophe to avoid selector parsing issues and font rendering differences.
	// e.g. "Your card in the making! We'll notify..." → "Your card in the making! We"
	normalizer := strings.NewReplacer("\u2018", "'", "\u2019", "'")
	matchText := normalizer.Replace(expectedMessage)
	if idx := strings.Index(matchText, "'"); idx > 0 {
		matchText = matchText[:idx]
	}

	// Use GetByText for reliable matching — avoids selector parsing issues
	snackbar := sc.page.GetByText(matchText, playwright.PageGetByTextOptions{Exact: playwright.Bool(false)})
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
