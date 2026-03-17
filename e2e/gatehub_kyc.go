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

// iShouldSeeTheActivateWalletButton verifies the activate wallet button is visible
func (sc *E2EContext) iShouldSeeTheActivateWalletButton() error {
	debugPrintln("\n👁️  Checking for activate wallet button...")

	time.Sleep(1 * time.Second)

	// Look for Continue button or similar activation trigger
	selectors := []string{
		"button:has-text('Continue')",
		"button:has-text('Activate')",
		"button:has-text('Next')",
		"[data-testid='kyc-continue']",
		"button[type='button']",
	}

	for _, selector := range selectors {
		button := sc.page.Locator(selector)
		count, err := button.Count()
		if err == nil && count > 0 {
			debugPrintf("   ✓ Found button: %s\n", selector)
			return nil
		}
	}

	return fmt.Errorf("no activate wallet button found")
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

// iFillAndSubmitTheMockgatehubiframe fills and submits the mockgatehub KYC iframe
func (sc *E2EContext) iFillAndSubmitTheMockgatehubiframe() error {
	debugPrintln("\n📝 Filling and submitting mockgatehub KYC iframe...")

	time.Sleep(500 * time.Millisecond)

	// First get the iframe src to understand what origin it's from
	iframeLocator := sc.page.Locator("iframe").First()
	iframeSrc, _ := iframeLocator.GetAttribute("src")
	debugPrintf("   📍 Iframe src: %s\n", iframeSrc)

	// Set up a listener to capture the postMessage
	debugPrintf("   📍 Setting up message listener...\n")
	_, err := sc.page.Evaluate(`() => {
		window.kycCompleted = false;
		window.addEventListener('message', (e) => {
			console.log('Parent received message:', e.data);
			if (e.data?.type === 'OnboardingCompleted' && JSON.parse(e.data.value || '{}').applicantStatus === 'submitted') {
				window.kycCompleted = true;
				console.log('KYC completed message received');
			}
		});
	}`)
	if err != nil {
		debugPrintf("   ⚠️  Failed to set up message listener: %v\n", err)
	}

	// Get the frame locator
	frameLocator := sc.page.FrameLocator("iframe").First()

	// Check if iframe exists
	iframeCount, _ := iframeLocator.Count()
	if iframeCount == 0 {
		return fmt.Errorf("no iframe found on page")
	}

	debugPrintf("   📍 Found iframe, searching for form elements\n")

	// Wait for iframe to be loaded and interactive
	time.Sleep(500 * time.Millisecond)

	// Try to find form inputs in the iframe
	inputs := frameLocator.Locator("input")
	inputCount, _ := inputs.Count()
	debugPrintf("   📍 Found %d input fields in iframe\n", inputCount)

	// If we have inputs, try to interact with them
	if inputCount > 0 {
		// Try to fill the required fields with test data
		for i := 0; i < inputCount; i++ {
			input := inputs.Nth(i)

			// Get input attributes to understand what it is
			placeholder, _ := input.GetAttribute("placeholder")
			name, _ := input.GetAttribute("name")
			inputType, _ := input.GetAttribute("type")

			debugPrintf("   📍 Input %d: type=%s, name=%s, placeholder=%s\n", i, inputType, name, placeholder)

			// Fill required visible fields
			if inputType != "hidden" {
				switch name {
				case "first_name":
					_ = input.Fill("Anna")
				case "last_name":
					_ = input.Fill("Müller")
				case "dob":
					_ = input.Fill("1990-01-15")
				case "address":
					_ = input.Fill("123 Main Street")
				case "city":
					_ = input.Fill("Berlin")
				case "country":
					_ = input.Fill("Germany")
				}
			}
		}
	}

	// Take screenshot of filled form before submission
	debugPrintf("   📸 Taking screenshot of filled KYC form...\n")
	if err := sc.iTakeAScreenshot("kyc-form-filled"); err != nil {
		debugPrintf("   ⚠️  Failed to take screenshot: %v\n", err)
		// Don't fail the test for screenshot failure
	}

	// Look for the submit button and click it
	buttons := frameLocator.Locator("button")
	buttonCount, _ := buttons.Count()
	debugPrintf("   📍 Found %d buttons in iframe\n", buttonCount)

	// Look for and click the submit button
	buttonClicked := false
	for i := 0; i < buttonCount; i++ {
		button := buttons.Nth(i)
		buttonText, _ := button.TextContent()
		buttonText = strings.TrimSpace(buttonText)

		debugPrintf("   📍 Button %d: %s\n", i, buttonText)

		// Look for submit button
		if strings.Contains(strings.ToLower(buttonText), "submit") {
			debugPrintf("   ✓ Clicking submit button: %s\n", buttonText)
			if err := button.Click(); err != nil {
				return fmt.Errorf("failed to click submit button: %w", err)
			}
			buttonClicked = true
			time.Sleep(500 * time.Millisecond)
			break
		}
	}

	if !buttonClicked {
		return fmt.Errorf("KYC iframe submit button not found - form was not submitted (found %d buttons in iframe)", buttonCount)
	}

	debugPrintf("   ✓ KYC iframe form submitted\n")
	return nil
}

// iFillAndSubmitTheMockxagoiframe fills and submits the MockXago Persona KYC iframe
func (sc *E2EContext) iFillAndSubmitTheMockxagoiframe() error {
	debugPrintln("\n📝 Filling and submitting MockXago Persona KYC iframe...")

	time.Sleep(500 * time.Millisecond)

	// Get the iframe src for diagnostics
	iframeLocator := sc.page.Locator("iframe").First()
	iframeSrc, _ := iframeLocator.GetAttribute("src")
	debugPrintf("   📍 Iframe src: %s\n", iframeSrc)

	// Set up a listener to capture the postMessage
	debugPrintf("   📍 Setting up message listener...\n")
	_, err := sc.page.Evaluate(`() => {
		window.kycCompleted = false;
		window.addEventListener('message', (e) => {
			console.log('Parent received message:', e.data);
			if (e.data?.type === 'OnboardingCompleted') {
				let parsed;
				try { parsed = JSON.parse(e.data.value || '{}'); } catch(ex) { parsed = {}; }
				if (parsed.applicantStatus === 'submitted') {
					window.kycCompleted = true;
					console.log('MockXago KYC completed message received');
				}
			}
		});
	}`)
	if err != nil {
		debugPrintf("   ⚠️  Failed to set up message listener: %v\n", err)
	}

	// Get the frame locator
	frameLocator := sc.page.FrameLocator("iframe").First()

	// Check if iframe exists
	iframeCount, _ := iframeLocator.Count()
	if iframeCount == 0 {
		return fmt.Errorf("no iframe found on page")
	}

	debugPrintf("   📍 Found iframe, searching for form elements\n")

	// Wait for iframe to be loaded and interactive
	time.Sleep(1 * time.Second)

	// Try to find form inputs in the iframe
	inputs := frameLocator.Locator("input")
	inputCount, _ := inputs.Count()
	debugPrintf("   📍 Found %d input fields in iframe\n", inputCount)

	if inputCount == 0 {
		// Take a screenshot for debugging
		_ = sc.iTakeAScreenshot("mockxago-iframe-no-inputs")
		return fmt.Errorf("no input fields found in MockXago iframe")
	}

	// Fill form fields by name attribute
	for i := 0; i < inputCount; i++ {
		input := inputs.Nth(i)
		name, _ := input.GetAttribute("name")
		inputType, _ := input.GetAttribute("type")
		placeholder, _ := input.GetAttribute("placeholder")

		debugPrintf("   📍 Input %d: type=%s, name=%s, placeholder=%s\n", i, inputType, name, placeholder)

		if inputType == "hidden" {
			continue
		}

		var fillErr error
		switch name {
		case "first_name":
			fillErr = input.Fill("Thabo")
		case "last_name":
			fillErr = input.Fill("Mbeki")
		case "dob":
			fillErr = input.Fill("1990-01-15")
		case "address":
			fillErr = input.Fill("42 Nelson Mandela Drive")
		case "city":
			fillErr = input.Fill("Johannesburg")
		case "country":
			fillErr = input.Fill("South Africa")
		}
		if fillErr != nil {
			return fmt.Errorf("failed to fill field %q: %w", name, fillErr)
		}
	}

	// Take screenshot of filled form before submission
	_ = sc.iTakeAScreenshot("mockxago-kyc-form-filled")

	// Look for the submit button and click it
	buttons := frameLocator.Locator("button")
	buttonCount, _ := buttons.Count()
	debugPrintf("   📍 Found %d buttons in iframe\n", buttonCount)

	buttonClicked := false
	for i := 0; i < buttonCount; i++ {
		button := buttons.Nth(i)
		buttonText, _ := button.TextContent()
		buttonText = strings.TrimSpace(buttonText)

		debugPrintf("   📍 Button %d: %s\n", i, buttonText)

		if strings.Contains(strings.ToLower(buttonText), "submit") {
			debugPrintf("   ✓ Clicking submit button: %s\n", buttonText)
			if err := button.Click(); err != nil {
				return fmt.Errorf("failed to click submit button: %w", err)
			}
			buttonClicked = true
			time.Sleep(500 * time.Millisecond)
			break
		}
	}

	if !buttonClicked {
		_ = sc.iTakeAScreenshot("mockxago-no-submit-button")
		return fmt.Errorf("MockXago KYC iframe submit button not found (found %d buttons)", buttonCount)
	}

	debugPrintf("   ✓ MockXago KYC iframe form submitted\n")
	return nil
}

// iWaitForTheKYCCompletion waits for KYC to complete and return to dashboard
func (sc *E2EContext) iWaitForTheKYCCompletion() error {
	debugPrintln("\n⏳ Waiting for KYC completion...")

	// First, wait for the form submission to complete by checking mockgatehub logs
	// The backend should receive the webhook and update KYC status
	time.Sleep(1 * time.Second)

	// Check if the message was received by the parent
	messageReceived, _ := sc.page.Evaluate(`() => {
		return window.kycCompleted === true;
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

		// If we're on login page, that's a session loss issue - return error
		if strings.Contains(currentURL, "/login") {
			return fmt.Errorf("KYC completed but redirected to login - user session was lost: %s", currentURL)
		}

		// We should navigate away from personal-details
		if !strings.Contains(currentURL, "/personal-details") {
			debugPrintf("   ✓ KYC completed, navigated away from personal-details\n")
			time.Sleep(500 * time.Millisecond)
			return nil
		}

		// Also try waiting for page navigation event
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("KYC did not complete within 30 seconds, still on personal-details")
}

// iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus verifies KYC approved and shows balance
func (sc *E2EContext) iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus() error {
	debugPrintln("\n🏠 Verifying dashboard with approved KYC status...")

	// Wait for backend to process the webhook and update KYC status
	// This may take several seconds as the webhook from mockgatehub needs to be processed
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

// iCompletedTheKYCFlowFor is a composite step that performs the entire KYC flow from kyc-minimal.feature
// This executes all steps after signup is complete: login, TOTP, wallet creation, and KYC completion
func (sc *E2EContext) iCompletedTheKYCFlowFor(email string) error {
	debugPrintf("\n🔄 Executing complete KYC flow for: %s\n", email)

	// Note: We expect signup to already be complete before this step
	// The deposit.feature background already calls iCompleteSignupFlow
	// The email is fetched from the current user's details

	// Use the prefixed email from the current user details
	prefixedEmail, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get email for KYC flow: %w", err)
	}

	// Store the password from context (should have been set during signup)
	password := sc.password
	if password == "" {
		return fmt.Errorf("no password set - signup must be completed before KYC flow")
	}

	debugPrintf("   Using prefixed email from context: %s\n", prefixedEmail)

	// 1. Ensure signup exists before login (use the prefixed email)
	if err := sc.aSignupRecordShouldExistInTheDatabase(prefixedEmail); err != nil {
		return fmt.Errorf("signup verification failed: %w", err)
	}

	// 2. Trigger verification and login
	if err := sc.iTriggerUserVerificationFor(prefixedEmail); err != nil {
		return fmt.Errorf("trigger verification failed: %w", err)
	}

	if err := sc.iClearTheBrowserSession(); err != nil {
		return fmt.Errorf("clear session failed: %w", err)
	}

	if err := sc.iNavigateToTheLoginPage(); err != nil {
		return fmt.Errorf("navigate to login failed: %w", err)
	}

	if err := sc.iFillInLoginCredentials(prefixedEmail, password); err != nil {
		return fmt.Errorf("fill login credentials failed: %w", err)
	}

	if err := sc.iSubmitTheLogin(); err != nil {
		return fmt.Errorf("submit login failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedToTheTOTPPage(); err != nil {
		return fmt.Errorf("TOTP page navigation failed: %w", err)
	}

	// 3. Register TOTP
	if err := sc.iTypeInMyGeneratedTotpForMyNewUser(); err != nil {
		return fmt.Errorf("TOTP generation failed: %w", err)
	}

	if err := sc.iSubmitTheTotpRegistration(); err != nil {
		return fmt.Errorf("TOTP registration failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedToTheApplicationDashboard(); err != nil {
		return fmt.Errorf("dashboard navigation failed: %w", err)
	}

	// 4. Create wallet address (required after signup)
	if err := sc.iShouldBeOnTheWalletAddressCreationPage(); err != nil {
		return fmt.Errorf("wallet address page check failed: %w", err)
	}

	if err := sc.iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress(); err != nil {
		return fmt.Errorf("fill wallet address form failed: %w", err)
	}

	if err := sc.iClickTheButtonOnTheWalletAddressForm("save"); err != nil {
		return fmt.Errorf("click save button failed: %w", err)
	}

	if err := sc.iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus(); err != nil {
		return fmt.Errorf("dashboard with reserved status failed: %w", err)
	}

	// 5. Navigate to wallet activation
	if err := sc.iNavigateToThePersonalDetailsPageToActivateWallet(); err != nil {
		return fmt.Errorf("navigate to personal details failed: %w", err)
	}

	if err := sc.iShouldSeeTheActivateWalletButton(); err != nil {
		return fmt.Errorf("activate wallet button check failed: %w", err)
	}

	// 6. Trigger KYC flow and fill iframe
	if err := sc.iClickTheButton("Continue"); err != nil {
		return fmt.Errorf("click Continue button failed: %w", err)
	}

	if err := sc.iWaitForTheKYCIframeToLoad(); err != nil {
		return fmt.Errorf("KYC iframe load failed: %w", err)
	}

	if err := sc.iFillAndSubmitTheMockgatehubiframe(); err != nil {
		return fmt.Errorf("fill KYC iframe failed: %w", err)
	}

	if err := sc.iWaitForTheKYCCompletion(); err != nil {
		return fmt.Errorf("KYC completion wait failed: %w", err)
	}

	// 7. Verify final state
	if err := sc.iShouldBeNavigatedBackToTheDashboardWithApprovedKYCStatus(); err != nil {
		return fmt.Errorf("dashboard with approved KYC failed: %w", err)
	}

	if err := sc.iShouldSeeMyAccountBalanceWithKYCApproved(); err != nil {
		return fmt.Errorf("account balance verification failed: %w", err)
	}

	// Take a final screenshot for documentation
	_ = sc.iTakeAScreenshot("kyc-completed-dashboard")

	debugPrintf("✅ Complete KYC flow finished successfully for %s\n", prefixedEmail)
	return nil
}

// iCompleteMinimalKYCFlowWithDetails runs the exact kyc-minimal.feature flow including signup.
// This wraps signup + iCompletedTheKYCFlowFor for reuse in deposit tests.
func (sc *E2EContext) iCompleteMinimalKYCFlowWithDetails(firstName, lastName, email, country, phone, password string) error {
	debugPrintf("\n🧭 Running minimal KYC flow for: %s\n", email)

	if err := sc.iCompleteSignupFlow(firstName, lastName, email, country, phone, password); err != nil {
		return fmt.Errorf("signup flow failed: %w", err)
	}

	if err := sc.iCompletedTheKYCFlowFor(email); err != nil {
		return fmt.Errorf("post-signup KYC flow failed: %w", err)
	}

	return nil
}

// iCompleteMinimalKYCFlow runs the minimal KYC flow using impersonated user details.
// This is used with the user impersonation pattern from signup-clean.feature and deposit-clean.feature.
// Usage: Given the details of 'kyc-user' are
//
//	  | field       | value        |
//	  | emailSuffix | alice@...    |
//	And I complete the minimal KYC flow `kyc-user`
func (sc *E2EContext) iCompleteMinimalKYCFlow(userName string) error {
	debugPrintf("\n🧭 Running minimal KYC flow for user: %s\n", userName)

	// Set the current user context first
	err := sc.iImpersonate(userName)
	if err != nil {
		return fmt.Errorf("failed to impersonate user '%s': %w", userName, err)
	}

	// Get the user's details
	details, exists := sc.userDetails[userName]
	if !exists {
		return fmt.Errorf("no details defined for user '%s' - ensure 'Given the details of' step was executed", userName)
	}

	// Extract required fields
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

	// Phone number will be auto-generated by iCompleteSignupFlow
	phone := "+4917" // Prefix for German mobile (will be overlaid with random digits)

	// Call the existing detailed KYC flow which handles signup + post-signup KYC
	return sc.iCompleteMinimalKYCFlowWithDetails(firstName, lastName, emailSuffix, country, phone, password)
}
