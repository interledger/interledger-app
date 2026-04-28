package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// thatMyFieldIs dynamically sets a field value for the current user
// Usage: Given that my "country" is "germany"
func (sc *E2EContext) thatMyFieldIs(fieldKey, value string) error {
	if sc.currentUser == "" {
		return fmt.Errorf("no user currently impersonated")
	}

	details, exists := sc.userDetails[sc.currentUser]
	if !exists {
		return fmt.Errorf("no details found for user '%s'", sc.currentUser)
	}

	// Store the field value
	details.Fields[fieldKey] = value

	// Also update the context fields if it's a standard field
	switch strings.ToLower(fieldKey) {
	case "country":
		sc.country = value
		debugPrintf("✓ Set country to: %s\n", value)
	case "firstname":
		sc.firstName = value
		debugPrintf("✓ Set firstName to: %s\n", value)
	case "lastname":
		sc.lastName = value
		debugPrintf("✓ Set lastName to: %s\n", value)
	case "dateofbirth", "date_of_birth":
		sc.dateOfBirth = value
		debugPrintf("✓ Set dateOfBirth to: %s\n", value)
	case "password":
		sc.password = value
		debugPrintf("✓ Set password\n")
	case "countrycode", "country_code":
		sc.countryCode = value
		debugPrintf("✓ Set countryCode to: %s\n", value)
	default:
		debugPrintf("✓ Set custom field '%s' to: %s\n", fieldKey, value)
	}

	return nil
}

// iCompletedTheAccountVerificationWorkflow triggers user verification
// Usage: And I completed the account verification workflow
func (sc *E2EContext) iCompletedTheAccountVerificationWorkflow() error {
	debugPrintln("\n🔄 Running account verification workflow...")

	if err := sc.iTriggerUserVerificationForMyself(); err != nil {
		return fmt.Errorf("failed to trigger user verification: %w", err)
	}

	debugPrintln("✓ Account verification workflow completed")
	return nil
}

// iFinishedTheTOTPRegistrationWorkflow completes login and TOTP registration
// Usage: And I finished the TOTP registration workflow
func (sc *E2EContext) iFinishedTheTOTPRegistrationWorkflow() error {
	debugPrintln("\n🔄 Running TOTP registration workflow...")

	// Clear session and navigate to login
	if err := sc.iClearTheBrowserSession(); err != nil {
		return fmt.Errorf("failed to clear browser session: %w", err)
	}

	if err := sc.iNavigateToTheLoginPage(); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Fill in login credentials
	if err := sc.iFillInTheLoginFormWithMyDetails(); err != nil {
		return fmt.Errorf("failed to fill login form: %w", err)
	}

	// Submit login
	if err := sc.iSubmitTheLogin(); err != nil {
		return fmt.Errorf("failed to submit login: %w", err)
	}

	// Verify we're on TOTP page
	if err := sc.iShouldBeNavigatedToTheTOTPPage(); err != nil {
		return fmt.Errorf("failed to navigate to TOTP page: %w", err)
	}

	// Enter TOTP code
	if err := sc.iTypeInMyGeneratedTotpForMyself(); err != nil {
		return fmt.Errorf("failed to generate and type TOTP: %w", err)
	}

	// Submit TOTP
	if err := sc.iSubmitTheTotpRegistration(); err != nil {
		return fmt.Errorf("failed to submit TOTP: %w", err)
	}

	currentURL := sc.page.URL()
	if strings.Contains(currentURL, "/phone-confirmation") {
		debugPrintln("   ✓ Reached phone confirmation page after TOTP")
	} else if err := sc.iShouldBeNavigatedToTheApplicationDashboard(); err != nil {
		return fmt.Errorf("failed to navigate to phone confirmation or dashboard: %w", err)
	}

	debugPrintln("✓ TOTP registration workflow completed")
	return nil
}

// iFinishedThePhoneConfirmationWorkflow completes the post-TOTP phone confirmation step.
// Usage: And I finished the phone confirmation workflow
func (sc *E2EContext) iFinishedThePhoneConfirmationWorkflow() error {
	debugPrintln("\n🔄 Running phone confirmation workflow...")

	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "/phone-confirmation") {
		return fmt.Errorf("not on phone confirmation page, current URL: %s", currentURL)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to resolve current user email: %w", err)
	}

	if err := sc.markKratosPhoneAsVerified(email); err != nil {
		return fmt.Errorf("failed to mark phone as verified: %w", err)
	}

	if _, err := sc.page.Goto(sc.baseURL+"/phone-confirmation", playwright.PageGotoOptions{
		Timeout:   playwright.Float(15000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("failed to reload phone confirmation page after verification: %w", err)
	}

	currentURL = sc.page.URL()
	if !strings.Contains(currentURL, "/wallet-address") && !strings.Contains(currentURL, "/dashboard") {
		return fmt.Errorf("expected redirect away from phone confirmation after verification, current URL: %s", currentURL)
	}

	debugPrintln("✓ Phone confirmation workflow completed")
	return nil
}

// iFinishedTheWalletAddressCreationWorkflow completes the wallet address creation flow
// Usage: And I finished the wallet address creation workflow
func (sc *E2EContext) iFinishedTheWalletAddressCreationWorkflow() error {
	debugPrintln("\n🔄 Running wallet address creation workflow...")

	// Verify we're on wallet address creation page.
	// In high-concurrency runs the wallet may already be auto-provisioned, in
	// which case we skip manual form submission and only verify reserved state.
	if err := sc.iShouldBeRedirectedToTheWalletAddressCreationPage(); err != nil {
		debugPrintf("   ⚠️  Wallet-address redirect not observed: %v\n", err)
		if dbErr := sc.waitForStableWalletCount(1, 2, 5*time.Second); dbErr == nil {
			debugPrintln("   ✓ Wallet already exists in DB; skipping manual wallet-address form")
		} else {
			return fmt.Errorf("not on wallet address creation page: %w", err)
		}
	} else {
		// Fill and submit the wallet address form
		if err := sc.iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress(); err != nil {
			return fmt.Errorf("failed to fill wallet address form: %w", err)
		}

		// Take screenshot for debugging
		if err := sc.iTakeAScreenshot("wallet-address-created"); err != nil {
			debugPrintf("⚠️  Failed to take screenshot: %v\n", err)
			// Don't fail the workflow for screenshot failure
		}

		// Click save button — retry once if a transient error (e.g. Bad Gateway) replaced the page
		if err := sc.iClickTheButtonOnTheWalletAddressForm("save"); err != nil {
			debugPrintf("⚠️  First save attempt failed: %v — reloading and retrying\n", err)
			sc.iTakeAScreenshot("wallet-address-save-retry")

			if _, reloadErr := sc.page.Reload(playwright.PageReloadOptions{
				Timeout: playwright.Float(15000),
			}); reloadErr != nil {
				return fmt.Errorf("failed to reload page after save failure: %w", reloadErr)
			}
			time.Sleep(2 * time.Second)

			// Re-fill form on the reloaded page
			if fillErr := sc.iFillInAndSubmitTheWalletAddressFormWithAUniqueAddress(); fillErr != nil {
				return fmt.Errorf("failed to refill wallet address form after reload: %w", fillErr)
			}

			if retryErr := sc.iClickTheButtonOnTheWalletAddressForm("save"); retryErr != nil {
				return fmt.Errorf("failed to click save button after retry: %w", retryErr)
			}
		}
	}

	// Verify we're back on dashboard with reserved status
	if err := sc.iShouldBeNavigatedBackToTheDashboardWithReservedWalletStatus(); err != nil {
		return fmt.Errorf("failed to navigate to dashboard with reserved status: %w", err)
	}

	// Take screenshot of reserved state
	if err := sc.iTakeAScreenshot("wallet-reserved-status"); err != nil {
		debugPrintf("⚠️  Failed to take screenshot: %v\n", err)
		// Don't fail the workflow for screenshot failure
	}

	debugPrintln("✓ Wallet address creation workflow completed")
	return nil
}

// iShouldBeShownTheActivateWalletPromptForm verifies the activate wallet prompt is shown
// Usage: Then I should be shown the "Activate wallet" prompt form
func (sc *E2EContext) iShouldBeShownTheActivateWalletPromptForm(promptText string) error {
	debugPrintf("\n👁️  Verifying '%s' prompt is shown...\n", promptText)

	// Navigate to personal details page
	if err := sc.iNavigateToThePersonalDetailsPageToActivateWallet(); err != nil {
		return fmt.Errorf("failed to navigate to personal details page: %w", err)
	}

	// South Africa/MockXago renders the iframe directly without a separate
	// "Continue" button prompt.
	if strings.EqualFold(sc.country, "south africa") {
		iframe := sc.page.Locator("iframe[title='Activate wallet'], iframe")
		if err := iframe.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("activate wallet prompt not visible for south africa flow: %w", err)
		}
	} else {
		// Verify the activate wallet button is visible for non-Xago providers.
		if err := sc.iShouldSeeTheActivateWalletButton(); err != nil {
			return fmt.Errorf("activate wallet button not visible: %w", err)
		}
	}

	// Take screenshot for debugging
	if err := sc.iTakeAScreenshot("activate-wallet-prompt"); err != nil {
		debugPrintf("⚠️  Failed to take screenshot: %v\n", err)
		// Don't fail for screenshot failure
	}

	debugPrintf("✓ '%s' prompt form is displayed\n", promptText)
	return nil
}
