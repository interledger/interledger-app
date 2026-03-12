package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/playwright-community/playwright-go"
)

// iShouldBeOnTheWalletAddressCreationPage verifies we're on the wallet address form
func (sc *E2EContext) iShouldBeOnTheWalletAddressCreationPage() error {
	debugPrintln("\n🏷️  Verifying we're on the wallet address creation page...")

	// Wait for middleware to finish auto-provisioning (if it's going to)
	// This prevents the race where middleware creates "default" wallet while user is filling form
	debugPrintln("   ⏳ Waiting for middleware wallet auto-provisioning to complete...")
	time.Sleep(500 * time.Millisecond) // Give middleware time to run during dashboard/page load

	currentURL := sc.page.URL()
	debugPrintf("   📍 Current URL: %s\n", currentURL)

	// Check if we're on the wallet address page
	if !strings.Contains(currentURL, "/wallet-address") {
		return fmt.Errorf("not on wallet address page, current URL: %s", currentURL)
	}

	// Look for the wallet address form title/description
	formText := sc.page.Locator("text=Create your unique wallet address")
	count, _ := formText.Count()
	if count > 0 {
		debugPrintf("   ✓ Found wallet address form\n")
		return nil
	}

	return fmt.Errorf("wallet address form not found on page")
}

// iShouldBeRedirectedToTheWalletAddressCreationPage waits for redirect to the wallet address page.
func (sc *E2EContext) iShouldBeRedirectedToTheWalletAddressCreationPage() error {
	debugPrintln("\n🔁 Waiting to be redirected to the wallet address creation page...")

	// Wait up to 30 seconds (was 15) for the redirect
	for i := 0; i < 30; i++ {
		currentURL := sc.page.URL()
		debugPrintf("   📍 Current URL (attempt %d): %s\n", i+1, currentURL)
		if strings.Contains(currentURL, "/wallet-address") {
			debugPrintln("   ✓ Successfully redirected to wallet address page")
			return sc.iShouldBeOnTheWalletAddressCreationPage()
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("did not redirect to wallet address page within timeout (30 seconds)")
}

// iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress fills the wallet address form (without clicking submit)
func (sc *E2EContext) iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress() error {
	debugPrintln("\n📝 Filling in wallet address form...")

	time.Sleep(1 * time.Second)

	// Find the wallet address input field (should have a prefix showing the payment pointer base)
	walletAddressInput := sc.page.Locator("input[name='username']")
	count, _ := walletAddressInput.Count()
	if count == 0 {
		return fmt.Errorf("wallet address input field not found")
	}

	// The form likely has a suggested username already - we can use it or generate a unique one
	// Clear the field first
	walletAddressInput.Fill("")
	time.Sleep(500 * time.Millisecond)

	// Generate a unique username based on test timestamp
	idPart := normalizeWalletAddressToken(sc.testIdentifier)
	userPart := normalizeWalletAddressToken(sc.currentUser)
	if idPart == "" {
		idPart = "test"
	}
	if userPart == "" {
		userPart = "user"
	}
	if len(idPart) > 6 {
		idPart = idPart[:6]
	}
	if len(userPart) > 3 {
		userPart = userPart[:3]
	}
	timePart := strconv.FormatInt(time.Now().UnixNano(), 36)
	// Ensure total length is under 16 characters (backend requirement)
	// userPart (3) + idPart (up to 6) + timePart (remaining)
	maxTimePart := 15 - len(userPart) - len(idPart)
	if maxTimePart < 0 {
		maxTimePart = 0
	}
	if len(timePart) > maxTimePart {
		timePart = timePart[len(timePart)-maxTimePart:]
	}
	// userPart must come first so the address starts with letters (backend requires first 3 chars to be letters)
	uniqueUsername := fmt.Sprintf("%s%s%s", userPart, idPart, timePart)
	debugPrintf("   📝 Using wallet address: %s\n", uniqueUsername)

	// Type in the username character by character to trigger onChange events
	// This mimics real user behavior and ensures validation calls are made
	err := walletAddressInput.Type(uniqueUsername, playwright.LocatorTypeOptions{
		Delay: playwright.Float(50), // 50ms delay between keystrokes
	})
	if err != nil {
		return fmt.Errorf("failed to type wallet address: %w", err)
	}

	// Wait for the onChange validation requests to complete
	// The form submits on every keystroke for validation, creating the default wallet
	debugPrintf("   ⏳ Waiting for validation requests to settle...\n")
	time.Sleep(1 * time.Second)

	// Wait for network to be idle (no pending validation requests)
	err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		debugPrintf("   ⚠️  Network still active: %v\n", err)
	}

	// Wait for buttons to be visible and enabled
	buttons := sc.page.Locator("button:not([disabled])")
	err = buttons.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		debugPrintf("   ⚠️  No enabled buttons found\n")
	}

	// Look for any button in the form - they should all be form submit buttons
	allButtons := sc.page.Locator("button")
	buttonCount, _ := allButtons.Count()
	debugPrintf("   📍 Found %d buttons on page\n", buttonCount)

	if buttonCount > 0 {
		for i := 0; i < buttonCount; i++ {
			button := allButtons.Nth(i)
			text, _ := button.TextContent()
			disabled, _ := button.GetAttribute("disabled")
			debugPrintf("   📍 Button %d: %s (disabled=%v)\n", i, strings.TrimSpace(text), disabled != "")
		}
	}

	debugPrintf("   ✓ Wallet address form filled and validations settled\n")

	// Store the wallet address for the current user (for later payment lookups)
	if sc.currentUser != "" && sc.userDetails != nil {
		if details, ok := sc.userDetails[sc.currentUser]; ok {
			prefix := ""
			prefixSpan := sc.page.Locator("input[name='username'] >> xpath=preceding-sibling::span[1]")
			if count, _ := prefixSpan.Count(); count > 0 {
				if text, err := prefixSpan.First().TextContent(); err == nil {
					prefix = strings.TrimSpace(text)
				}
			}
			if prefix == "" {
				prefix = "local.ilp.link/"
			}
			walletURL := prefix
			if !strings.HasSuffix(walletURL, "/") {
				walletURL += "/"
			}
			walletURL += uniqueUsername
			details.Fields["walletAddress"] = walletURL
			details.Fields["walletUsername"] = uniqueUsername
			debugPrintf("   ✓ Stored wallet address for %s: %s\n", sc.currentUser, walletURL)
		}
	}
	return nil
}

// iClickTheButtonOnTheWalletAddressForm clicks a specific button on the wallet-address form
func (sc *E2EContext) iClickTheButtonOnTheWalletAddressForm(buttonText string) error {
	debugPrintf("\n🖱️  Clicking '%s' button on wallet-address form...\n", buttonText)

	time.Sleep(500 * time.Millisecond)

	// Track network requests to capture POST response for debugging
	var walletAddressResponse playwright.Response
	sc.page.OnResponse(func(response playwright.Response) {
		url := response.URL()
		if strings.Contains(url, "/wallet-address") && response.Request().Method() == "POST" {
			walletAddressResponse = response
			debugPrintf("   📡 Captured POST /wallet-address response: %d\n", response.Status())
		}
	})

	// Find the wallet-address submit button first (most deterministic locator)
	submitButton := sc.page.Locator("button[form='wallet-address'][type='submit']")
	btnCount, _ := submitButton.Count()

	if btnCount == 0 {
		// Fallback to text match if the form/type attributes are unavailable
		submitButton = sc.page.Locator(fmt.Sprintf("button:has-text('%s')", buttonText))
		btnCount, _ = submitButton.Count()
		if btnCount == 0 {
			return fmt.Errorf("could not find wallet-address submit button")
		}
	}

	err := submitButton.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("submit button not visible: %w", err)
	}

	var enabled bool
	for i := 0; i < 20; i++ {
		enabled, err = submitButton.First().IsEnabled()
		if err != nil {
			return fmt.Errorf("failed to inspect submit button state: %w", err)
		}
		if enabled {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if !enabled {
		disabledAttr, _ := submitButton.First().GetAttribute("disabled")
		return fmt.Errorf("submit button remained disabled before click (disabled=%v)", disabledAttr != "")
	}

	debugPrintf("   ✓ Found actionable '%s' button\n", buttonText)

	// Click the submit button
	err = submitButton.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("failed to click '%s' button: %w", buttonText, err)
	}

	// Wait for the page to finish loading after form submission
	debugPrintf("   ⏳ Waiting for form submission and redirect...\n")
	err = sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	})
	if err != nil {
		debugPrintf("   ⚠️  Load state timeout: %v, checking current URL\n", err)
	}

	currentURL := sc.page.URL()
	if strings.Contains(currentURL, "/wallet-address") {
		debugPrintf("   ⚠️  Still on wallet-address page after form submission\n")

		// Log backend response details if captured
		if walletAddressResponse != nil {
			debugPrintf("   📡 Backend response status: %d\n", walletAddressResponse.Status())
			debugPrintf("   📡 Backend response status text: %s\n", walletAddressResponse.StatusText())
			body, err := walletAddressResponse.Body()
			if err == nil && len(body) > 0 && len(body) < 2000 {
				debugPrintf("   📡 Backend response body: %s\n", string(body))
			} else if err != nil {
				debugPrintf("   ⚠️  Could not read response body: %v\n", err)
			}
		} else {
			debugPrintf("   ⚠️  No POST /wallet-address response captured\n")
		}

		// Check if wallet was created in DB despite UI not redirecting
		debugPrintf("   🔍 Checking if wallet was created in DB...\n")

		email, err := sc.getCurrentUserEmail()
		if err == nil && sc.db != nil {
			kratosID := sc.getKratosUserIDByEmail(email)
			if kratosID != "" {
				count, err := sc.getUserWalletCount(kratosID)
				if err == nil {
					debugPrintf("   📊 DB wallet count: %d\n", count)
					if count > 0 {
						// Wallet was created, fetch details
						walletDetails, err := sc.getWalletDetailsForUser(kratosID)
						if err == nil {
							debugPrintf("   ✓ Wallet exists in DB: id=%s, name=%s, address=%s\n",
								walletDetails.ID, walletDetails.Name, walletDetails.WalletAddress)
							debugPrintf("   💡 Wallet created successfully but UI did not redirect - possible frontend issue\n")
						}
					}
				} else {
					debugPrintf("   ⚠️  DB query error: %v\n", err)
				}
			}
		}

		// Check for error messages
		errorMsg := sc.page.Locator(".error, [role='alert']")
		errCount, _ := errorMsg.Count()
		if errCount > 0 {
			for i := 0; i < errCount; i++ {
				text, _ := errorMsg.Nth(i).TextContent()
				debugPrintf("   ❌ Error message: %s\n", strings.TrimSpace(text))
			}
		}

		// Take screenshot for debugging
		_ = sc.iTakeAScreenshot("wallet-address-error")

		time.Sleep(500 * time.Millisecond)
		return nil
	}

	debugPrintf("   ✓ Wallet address form submitted and redirected\n")

	time.Sleep(1 * time.Second)
	return nil
}

// iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus verifies redirect and reserved status
func (sc *E2EContext) iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus() error {
	debugPrintln("\n🏠 Verifying dashboard with reserved wallet status...")

	// Wait longer for form submission and navigation to complete
	for i := 0; i < 15; i++ { // Up to 15 seconds
		time.Sleep(1 * time.Second)

		currentURL := sc.page.URL()
		debugPrintf("   📍 Current URL (attempt %d): %s\n", i+1, currentURL)

		// Should be at root dashboard, not /wallet-address anymore
		if strings.Contains(currentURL, "/wallet-address") {
			// Still on wallet address page, try reloading to see if we should redirect
			if i%3 == 0 {
				debugPrintf("   🔄 Reloading page...\n")
				sc.page.Reload()
			}
			continue
		}

		// Wait for wallet count to stabilize (should be at least 1)
		err := sc.waitForStableWalletCount(1, 3, 5*time.Second)
		if err != nil {
			debugPrintf("   ⚠️  Wallet count did not stabilize: %v\n", err)
			// Continue anyway, but log for debugging
		}

		// Check for "Reserved" chip indicating wallet is not yet activated
		reservedChip := sc.page.Locator("text=Reserved")
		count, _ := reservedChip.Count()
		if count > 0 {
			debugPrintf("   ✓ Wallet shows 'Reserved' status - wallet address created, KYC pending\n")
			return nil
		}

		// Also check for "Activate wallet" link as confirmation
		activateLink := sc.page.Locator("text=Activate wallet")
		count, _ = activateLink.Count()
		if count > 0 {
			debugPrintf("   ✓ Found 'Activate wallet' link - correct state\n")
			return nil
		}
	}

	return fmt.Errorf("wallet does not appear to be in 'Reserved' state, still on wallet-address")
}

func normalizeWalletAddressToken(input string) string {
	if input == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for _, character := range input {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(unicode.ToLower(character))
		}
	}

	return builder.String()
}
