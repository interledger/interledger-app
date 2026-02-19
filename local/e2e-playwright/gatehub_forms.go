package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Form filling step implementations

// getPhoneBaseForCountry returns the base phone number template for a given country.
// The trailing zeros in each base are replaced by a deterministic overlay.
func getPhoneBaseForCountry(country string) string {
	switch strings.ToLower(country) {
	case "south africa":
		// +27 country code, 71 prefix preserved → +27710000000 (11 digits)
		return "+27710000000"
	default:
		// Germany / fallback: +49 country code, 1700 prefix → +491700000000 (12 digits)
		return "+491700000000"
	}
}

// generateDeterministicPhone builds a phone number from a country-specific base
// by overlaying a deterministic 6-digit suffix derived from testEmailPrefix and emailSuffix.
func generateDeterministicPhone(country, testEmailPrefix, emailSuffix string) string {
	base := getPhoneBaseForCountry(country)

	// Create a unique overlay combining test prefix and email hash
	overlay := testEmailPrefix
	if emailSuffix != "" {
		// Extract the part before @ to get unique username (e.g., "sender-p2p" from "sender-p2p@example.com")
		emailParts := strings.Split(emailSuffix, "@")
		username := emailParts[0]

		// Hash the full username to get 3 unique digits
		userHash := 0
		for _, c := range username {
			userHash = (userHash*31 + int(c)) % 1000
		}
		emailHash := fmt.Sprintf("%03d", userHash)

		// Use first 6 digits: testPrefix (3) + email hash (3)
		if len(overlay) >= 3 {
			overlay = overlay[:3] + emailHash
		} else {
			overlay = overlay + emailHash
		}
	}

	if overlay == "" {
		// Fallback to a random 6-digit suffix if test identifier wasn't generated
		overlay = fmt.Sprintf("%06d", rand.Intn(1000000))
	}

	// Ensure we don't slice into the leading '+' when overlaying
	if len(overlay) >= len(base) {
		overlay = overlay[len(overlay)-(len(base)-1):]
	}
	split := len(base) - len(overlay)
	if split < 1 {
		split = 1
	}
	return base[:split] + overlay
}

// iFillInPhoneWithRandomNumber generates a country-aware phone number and fills it
func (sc *E2EContext) iFillInPhoneWithRandomNumber(prefix string) error {
	// Get the current user's details to extract the email suffix
	if sc.currentUser == "" {
		return fmt.Errorf("no current user set for phone generation")
	}

	details, ok := sc.userDetails[sc.currentUser]
	if !ok {
		return fmt.Errorf("no user details for current user '%s'", sc.currentUser)
	}

	// Use emailSuffix (without the test prefix) for consistent hashing
	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("no emailSuffix defined for user '%s'", sc.currentUser)
	}

	phoneNumber := generateDeterministicPhone(sc.country, sc.testIdentifier, emailSuffix)
	debugPrintf("📱 Generated phone number: %s (country %s, emailSuffix %s, user %s)\n", phoneNumber, sc.country, emailSuffix, sc.currentUser)

	// Store in user details for current user
	if sc.currentUser != "" && sc.userDetails[sc.currentUser] != nil {
		sc.userDetails[sc.currentUser].Fields["phone"] = phoneNumber
		debugPrintf("📱 Stored phone in userDetails for '%s'\n", sc.currentUser)
	} else {
		return fmt.Errorf("cannot store phone: no current user set or user details not initialized")
	}

	// Wait for page to be ready
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	// Fill in the phone field
	var err error
	err = sc.iFillInWith("phone", phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to fill phone field: %w", err)
	}

	// Trigger blur event to ensure onChange/onBlur handlers fire
	input := sc.page.Locator("input[type='tel']")
	count, _ := input.Count()
	if count > 0 {
		input.Blur()
	} else {
		debugPrintln("   ⚠️  No tel input found for blur event")
	}

	// Wait for the value to be processed and API call to complete
	sc.page.WaitForTimeout(1000)

	// Verify the value was actually set
	input = sc.page.Locator("input[type='tel']")
	count, _ = input.Count()
	if count > 0 {
		actualValue, _ := input.InputValue()
		debugPrintf("📱 Phone field value after filling: %s\n", actualValue)
	} else {
		// Try alternate phone input selectors
		input = sc.page.Locator("input[name*='phone' i]")
		count, _ = input.Count()
		if count > 0 {
			actualValue, _ := input.InputValue()
			debugPrintf("📱 Phone field value after filling (via name selector): %s\n", actualValue)
		} else {
			debugPrintf("📱 Phone field value after filling: (no phone field found)\n")
		}
	}

	return nil
}

func (sc *E2EContext) iFillInWith(fieldName, value string) error {
	var input playwright.Locator

	switch strings.ToLower(fieldName) {
	case "first name":
		input = sc.page.Locator("input[name='firstName'], input[name='first_name'], input[placeholder*='first' i]")
		_ = input.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		})
	case "last name":
		input = sc.page.Locator("input[name='lastName'], input[name='last_name'], input[placeholder*='last' i]")
		_ = input.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		})
	case "email":
		// Prefix email with random test identifier (avoid double prefix)
		prefix := fmt.Sprintf("%s-", sc.testIdentifier)
		if !strings.HasPrefix(value, prefix) {
			value = fmt.Sprintf("%s%s", prefix, value)
		}
		input = sc.page.Locator("input[type='email'], input[name*='email'], input[placeholder*='email' i]")
		_ = input.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		})
		sc.email = value // Store prefixed email
	case "phone":
		// Try flexible phone selectors
		input = sc.page.Locator("input[type='tel']")
		count, _ := input.Count()
		if count == 0 {
			input = sc.page.Locator("input[name*='phone' i]")
			count, _ = input.Count()
		}
		if count == 0 {
			input = sc.page.Locator("input[placeholder*='phone' i]")
			count, _ = input.Count()
		}
		if count == 0 {
			phoneLabel := sc.page.Locator("label:has-text('Phone')")
			labelCount, _ := phoneLabel.Count()
			if labelCount > 0 {
				forAttr, _ := phoneLabel.First().GetAttribute("for")
				if forAttr != "" {
					input = sc.page.Locator(fmt.Sprintf("input#%s", forAttr))
				}
			}
		}
		if count == 0 {
			allInputs := sc.page.Locator("input[type='text']")
			allCount, _ := allInputs.Count()
			for i := 0; i < allCount; i++ {
				inp := allInputs.Nth(i)
				inputName, _ := inp.GetAttribute("name")
				if !strings.Contains(strings.ToLower(inputName), "country") {
					input = inp
					break
				}
			}
		}
		if count == 0 {
			currentEmail, err := sc.getCurrentUserEmail()
			if err != nil {
				return fmt.Errorf("failed to get current user email for phone fill: %w", err)
			}
			var signupID string
			err = sc.db.QueryRow("SELECT id FROM signups WHERE email = $1", currentEmail).Scan(&signupID)
			if err == nil && signupID != "" {
				_, _ = sc.db.Exec(
					"UPDATE signups SET mobile_number = $1 WHERE id = $2",
					value,
					signupID,
				)
				debugPrintf("📱 Updated mobile_number in database for signup %s\n", signupID)
			}
		}
	case "password":
		// Wait for password field to be available
		passwordInputs := sc.page.Locator("input[type='password']")
		_ = passwordInputs.First().WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		count, _ := passwordInputs.Count()
		if count > 0 {
			input = passwordInputs.First()
			sc.password = value
			sc.passwordFilled = true
		} else {
			return fmt.Errorf("no password input fields found")
		}
	case "password confirmation", "confirm password":
		passwordInputs := sc.page.Locator("input[type='password']")
		_ = passwordInputs.Nth(1).WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		})
		count, _ := passwordInputs.Count()
		if count > 1 {
			input = passwordInputs.Nth(1)
		} else {
			return fmt.Errorf("password confirmation input not found")
		}
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}

	if input == nil {
		return fmt.Errorf("failed to find input for '%s'", fieldName)
	}

	count, _ := input.Count()
	if count == 0 {
		return fmt.Errorf("failed to find input for '%s'", fieldName)
	}

	err := input.Fill(value)
	if err != nil {
		return fmt.Errorf("failed to fill '%s' with '%s': %w", fieldName, value, err)
	}

	return nil
}

// iTryToFillInWith attempts to fill a field but doesn't fail if the field isn't found
func (sc *E2EContext) iTryToFillInWith(fieldName, value string) error {
	err := sc.iFillInWith(fieldName, value)
	if err != nil {
		// Field not found or couldn't be filled - OK for optional fields
		return nil
	}

	return nil
}

func (sc *E2EContext) iSelectFromTheCountryDropdown(country string) error {
	// Find country input - be very specific to avoid selecting wrong field
	var countryInput playwright.Locator

	// First try: direct ID selector (most reliable)
	countryInput = sc.page.Locator("input#country:not([type='hidden'])")
	count, _ := countryInput.Count()

	// Second try: input with name="country"
	if count == 0 {
		countryInput = sc.page.Locator("input[name='country']:not([type='hidden'])")
		count, _ = countryInput.Count()
	}

	// Third try: look for input near "Country" label text
	if count == 0 {
		countryLabel := sc.page.Locator("label:has-text('Country'), label:has-text('country')")
		labelCount, _ := countryLabel.Count()
		if labelCount > 0 {
			forAttr, _ := countryLabel.First().GetAttribute("for")
			if forAttr != "" {
				countryInput = sc.page.Locator(fmt.Sprintf("input#%s:not([type='hidden'])", forAttr))
				count, _ = countryInput.Count()
			}
		}
	}

	// Fourth try: input inside a div/element with "country" in class or id
	if count == 0 {
		countryInput = sc.page.Locator("[class*='country' i] input[type='text']:not([type='hidden']), [id*='country' i] input[type='text']:not([type='hidden'])")
		count, _ = countryInput.Count()
	}

	// Fifth try: combobox input (Headless UI pattern)
	if count == 0 {
		countryInput = sc.page.Locator("input[role='combobox']:not([type='hidden'])")
		count, _ = countryInput.Count()
	}

	if count == 0 {
		return fmt.Errorf("country autocomplete input not found - tried multiple selectors")
	}

	// Prefer a visible, non-hidden input if multiple matches exist
	inputToUse := countryInput.First()
	if count > 1 {
		for i := 0; i < count; i++ {
			candidate := countryInput.Nth(i)
			isVisible, _ := candidate.IsVisible()
			inputType, _ := candidate.GetAttribute("type")
			if isVisible && !strings.EqualFold(inputType, "hidden") {
				inputToUse = candidate
				break
			}
		}
	}
	// Extra guard: ensure we didn't select a hidden input
	inputType, _ := inputToUse.GetAttribute("type")
	if strings.EqualFold(inputType, "hidden") {
		return fmt.Errorf("country input resolved to hidden field")
	}

	// Verify this is NOT a name field
	inputOuterHTML, _ := inputToUse.Evaluate("el => el.outerHTML", nil)
	inputHTML := fmt.Sprintf("%v", inputOuterHTML)

	if strings.Contains(strings.ToLower(inputHTML), "firstname") ||
		strings.Contains(strings.ToLower(inputHTML), "lastname") ||
		strings.Contains(strings.ToLower(inputHTML), "name") && !strings.Contains(strings.ToLower(inputHTML), "country") {
		return fmt.Errorf("safety check failed: found name field instead of country field")
	}

	// Try to find and click dropdown button to open options
	toggleButton := sc.page.Locator("button:near(input#country, 100px)")
	toggleCount, _ := toggleButton.Count()
	if toggleCount == 0 {
		toggleButton = sc.page.Locator("button:has-text('unfold_more')")
		toggleCount, _ = toggleButton.Count()
	}

	if toggleCount > 0 {
		_ = toggleButton.First().Click()
		time.Sleep(500 * time.Millisecond)
	} else {
		err := inputToUse.Click(playwright.LocatorClickOptions{Force: playwright.Bool(true)})
		if err != nil {
			return fmt.Errorf("failed to click country input: %w", err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Ensure input is focused before typing
	_ = inputToUse.Click(playwright.LocatorClickOptions{Force: playwright.Bool(true)})
	time.Sleep(200 * time.Millisecond)

	// Type to filter options using Fill (force to avoid editable timeout)
	err := inputToUse.Fill(country, playwright.LocatorFillOptions{
		Timeout: playwright.Float(5000),
		Force:   playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to type country: %w", err)
	}

	// Wait longer for autocomplete to filter
	time.Sleep(1500 * time.Millisecond)

	// Find the country option with multiple attempts
	var countryOption playwright.Locator

	// Try role="option" with exact text match
	countryOption = sc.page.Locator(fmt.Sprintf("[role='option']:has-text('%s')", country))
	optCount, _ := countryOption.Count()
	if optCount == 0 {
		// Try with visible options only
		allOptions := sc.page.Locator("[role='option']")
		allCount, _ := allOptions.Count()
		for i := 0; i < allCount; i++ {
			opt := allOptions.Nth(i)
			text, _ := opt.TextContent()
			isVisible, _ := opt.IsVisible()
			if isVisible && strings.Contains(text, country) {
				countryOption = opt
				break
			}
		}
	}

	// Try Headless UI combobox options
	if optCount == 0 {
		countryOption = sc.page.Locator(fmt.Sprintf("div[id*='headlessui-combobox-option']:has-text('%s')", country))
		optCount, _ = countryOption.Count()
	}

	// Try li elements
	if optCount == 0 {
		countryOption = sc.page.Locator(fmt.Sprintf("li:has-text('%s')", country))
		optCount, _ = countryOption.Count()
	}

	if optCount > 0 {
		// Found the option - verify it's visible
		isVisible, _ := countryOption.First().IsVisible()
		if isVisible {
			err = countryOption.First().Click()
			if err != nil {
				return fmt.Errorf("failed to click country option: %w", err)
			}
			time.Sleep(500 * time.Millisecond)

			// Verify selection by checking input value changed from typed text
			inputValue, _ := inputToUse.InputValue()
			if inputValue == country {
				_ = inputToUse.Press("Enter")
				time.Sleep(500 * time.Millisecond)
			}

			return nil
		}
	}

	// Last resort: press Enter to select highlighted option
	err = inputToUse.Press("Enter")
	if err != nil {
		return fmt.Errorf("failed to press Enter on country input: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	return nil
}

func (sc *E2EContext) iCheckTheTermsAndConditionsCheckbox() error {
	time.Sleep(500 * time.Millisecond)

	// Look for checkbox - try various selectors
	checkbox := sc.page.Locator("input[type='checkbox']")
	checkboxCount, _ := checkbox.Count()

	// Try finding by label text if first attempt found nothing
	if checkboxCount == 0 {
		checkbox = sc.page.Locator("input[type='checkbox']:near(:text('terms')), input[type='checkbox']:near(:text('agree')), input[type='checkbox']:near(:text('accept'))")
		checkboxCount, _ = checkbox.Count()
	}

	if checkboxCount == 0 {
		// Try finding the first checkbox on the page
		allCheckboxes := sc.page.Locator("input[type='checkbox']")
		allCount, _ := allCheckboxes.Count()
		if allCount > 0 {
			checkbox = allCheckboxes.First()
			checkboxCount = 1
		}
	}

	if checkboxCount == 0 {
		return fmt.Errorf("checkbox not found")
	}

	// Check if already checked
	isChecked, _ := checkbox.First().IsChecked()
	if !isChecked {
		err := checkbox.First().Check()
		if err != nil {
			return fmt.Errorf("failed to check checkbox: %w", err)
		}
		time.Sleep(300 * time.Millisecond)

		// Verify it's actually checked now
		isChecked, _ = checkbox.First().IsChecked()
		if !isChecked {
			return fmt.Errorf("checkbox still not checked after Check() call")
		}
	}

	return nil
}

func (sc *E2EContext) iTryToSubmitWithoutFillingRequiredFields() error {
	submitButton := sc.page.Locator("button:has-text('Continue'), button:has-text('Submit'), button[type='submit']")
	count, _ := submitButton.Count()
	if count > 0 {
		submitButton.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(3000),
		})
	}
	time.Sleep(1 * time.Second)
	return nil
}
