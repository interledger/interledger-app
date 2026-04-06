package main

import (
	"fmt"
	"strings"
	"time"
)

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
