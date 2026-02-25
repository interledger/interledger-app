package main

import (
	"github.com/playwright-community/playwright-go"
)

// iCompleteSignupFlow runs the full signup flow with provided user details.
func (sc *E2EContext) iCompleteSignupFlow(firstName, lastName, email, country, phone, password string) error {
	// Store password (email prefixing is handled by iFillInWith)
	sc.password = password
	sc.firstName = firstName
	sc.lastName = lastName
	sc.country = country

	if err := sc.iNavigateToTheSignupPage(); err != nil {
		return err
	}

	debugPrintf("   📍 Current URL after navigate: %s\n", sc.page.URL())

	// Removed: iClickTheButton("Sign Up") - now we navigate directly to /signup
	// Try clicking "Get started" or "Let's get started" button if it exists
	// Use a short timeout since it might not exist
	getStartedSelector := "button:has-text('Get started'), button:has-text('Let'), button[data-testid='signup-get-started']"
	if err := sc.page.Locator(getStartedSelector).First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(2000),
	}); err == nil {
		// Button was clicked successfully, wait for form to appear
		sc.page.WaitForTimeout(1000)
		debugPrintf("   📍 URL after Get Started: %s\n", sc.page.URL())
	}
	// else: button didn't exist or timed out, continue anyway

	if err := sc.iShouldSeeTheSignupForm(); err != nil {
		debugPrintln("   ⚠️  Form not visible on initial check, taking screenshot")
		_ = sc.iTakeAScreenshot("form-not-visible")
		return err
	}

	debugPrintf("   📍 Form is visible, starting to fill\n")

	if err := sc.iFillInWith("first name", firstName); err != nil {
		return err
	}
	if err := sc.iFillInWith("last name", lastName); err != nil {
		return err
	}
	// Pass ORIGINAL email - iFillInWith will prefix it
	if err := sc.iFillInWith("email", email); err != nil {
		return err
	}
	if err := sc.iSelectFromTheCountryDropdown(country); err != nil {
		return err
	}

	debugPrintln("   📝 Form filled, clicking Continue...")
	if err := sc.iClickTheButton("Continue"); err != nil {
		return err
	}

	debugPrintf("   📍 Current URL after Continue step 1: %s\n", sc.page.URL())

	if err := sc.iShouldBeOnStep(2); err != nil {
		return err
	}

	// Generate random phone number with the provided prefix (e.g., "+49")
	if err := sc.iFillInPhoneWithRandomNumber(phone); err != nil {
		return err
	}

	// Wait for phone input to be processed and state to update
	sc.page.WaitForTimeout(500)

	if err := sc.iClickTheButton("Continue"); err != nil {
		return err
	}

	// Fill password fields once (removed duplicate fills that caused 8s delay)
	if err := sc.iTryToFillInWith("password", password); err != nil {
		return err
	}
	if err := sc.iTryToFillInWith("password confirmation", password); err != nil {
		return err
	}
	if err := sc.iCheckTheTermsAndConditionsCheckbox(); err != nil {
		return err
	}
	if err := sc.iClickTheButton("Confirm"); err != nil {
		return err
	}

	return sc.theSignupShouldBeSubmitted()
}
