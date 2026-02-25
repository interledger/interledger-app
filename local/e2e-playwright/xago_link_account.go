package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iNavigateToTheDashboardPage navigates to a named dashboard page.
// "Home" → /, "Accounts" → /accounts, etc.
func (sc *E2EContext) iNavigateToTheDashboardPage(page string) error {
	debugPrintf("\n📊 Navigating to dashboard page: %s\n", page)

	var path string
	switch strings.ToLower(page) {
	case "home":
		path = "/"
	case "accounts":
		path = "/accounts"
	case "payments":
		path = "/payments"
	case "settings":
		path = "/settings"
	default:
		path = "/" + strings.ToLower(page)
	}

	url := sc.baseURL + path
	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(15000),
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to dashboard %s: %w", page, err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Navigated to dashboard %s: %s\n", page, url)
	return nil
}

// iPressOn clicks on any clickable element (link or button) with the given text.
// It iterates through all matches and clicks the first visible one.
func (sc *E2EContext) iPressOn(text string) error {
	debugPrintf("\n👆 Pressing on: %s\n", text)

	_ = sc.iTakeAScreenshot(fmt.Sprintf("before-press-%s", strings.ReplaceAll(strings.ToLower(text), " ", "-")))

	debugPrintf("   📍 Current URL before press: %s\n", sc.page.URL())

	// Try link first, then button
	selectors := []string{
		fmt.Sprintf("a:has-text('%s')", text),
		fmt.Sprintf("button:has-text('%s')", text),
		fmt.Sprintf("[role='link']:has-text('%s')", text),
		fmt.Sprintf("[role='button']:has-text('%s')", text),
	}

	for _, selector := range selectors {
		allMatches := sc.page.Locator(selector)
		count, _ := allMatches.Count()
		if count == 0 {
			continue
		}
		debugPrintf("   📍 Found %d element(s) with selector: %s\n", count, selector)

		// Iterate through all matches and click the first VISIBLE one
		for j := 0; j < count; j++ {
			el := allMatches.Nth(j)
			visible, _ := el.IsVisible()
			debugPrintf("   📍 Element %d visible=%v\n", j, visible)
			if !visible {
				continue
			}

			err := el.Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(10000),
			})
			if err == nil {
				sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State:   playwright.LoadStateNetworkidle,
					Timeout: playwright.Float(15000),
				})
				time.Sleep(500 * time.Millisecond)

				debugPrintf("   📍 Current URL after press: %s\n", sc.page.URL())
				debugPrintf("✓ Pressed on: %s\n", text)
				return nil
			}
			debugPrintf("   ⚠️  Click failed for element %d: %v\n", j, err)
		}
	}

	return fmt.Errorf("could not find clickable element with text '%s'", text)
}

// iShouldSeeTheForm verifies a form with the given title is visible on the page.
func (sc *E2EContext) iShouldSeeTheForm(formTitle string) error {
	debugPrintf("\n📋 Verifying form: %s\n", formTitle)

	// Wait for page to be ready
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(500 * time.Millisecond)

	_ = sc.iTakeAScreenshot(fmt.Sprintf("form-%s", strings.ReplaceAll(strings.ToLower(formTitle), " ", "-")))

	// Check for the form title text on the page
	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, formTitle) {
		debugPrintf("✓ Found form: %s\n", formTitle)
		return nil
	}

	// Also check page title (h1, h2)
	headings := sc.page.Locator("h1, h2, h3, [class*='title']")
	count, _ := headings.Count()
	for i := 0; i < count; i++ {
		text, _ := headings.Nth(i).TextContent()
		if strings.Contains(text, formTitle) {
			debugPrintf("✓ Found form title in heading: %s\n", formTitle)
			return nil
		}
	}

	return fmt.Errorf("form with title '%s' not found on page", formTitle)
}

// iFillInFieldWith fills a form field by label or name with the given value.
// This is a generic form fill step that locates inputs by label text, id, or name.
func (sc *E2EContext) iFillInFieldWith(fieldLabel, value string) error {
	debugPrintf("\n📝 Filling field '%s' with '%s'\n", fieldLabel, value)

	_ = sc.iTakeAScreenshot(fmt.Sprintf("fill-field-%s", strings.ReplaceAll(strings.ToLower(fieldLabel), " ", "-")))

	// Debug: log current URL and count all inputs
	debugPrintf("   📍 Current URL: %s\n", sc.page.URL())
	allInputs := sc.page.Locator("input")
	inputCount, _ := allInputs.Count()
	debugPrintf("   📍 Total inputs on page: %d\n", inputCount)
	for i := 0; i < inputCount && i < 10; i++ {
		inp := allInputs.Nth(i)
		id, _ := inp.GetAttribute("id")
		name, _ := inp.GetAttribute("name")
		inputType, _ := inp.GetAttribute("type")
		debugPrintf("   📍 Input %d: id=%s, name=%s, type=%s\n", i, id, name, inputType)
	}

	// Strategy 1: Find input directly by ID derived from label (camelCase)
	camelID := toCamelCase(fieldLabel)
	input := sc.page.Locator(fmt.Sprintf("input#%s", camelID))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via id '%s': %s\n", fieldLabel, camelID, value)
			return nil
		}
	}

	// Strategy 2: Find input by name attribute
	input = sc.page.Locator(fmt.Sprintf("input[name='%s']", camelID))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via name '%s': %s\n", fieldLabel, camelID, value)
			return nil
		}
	}

	// Strategy 3: Find input by associated label text (case-insensitive)
	label := sc.page.Locator(fmt.Sprintf("label:has-text('%s')", fieldLabel))
	if count, _ := label.Count(); count > 0 {
		forAttr, _ := label.First().GetAttribute("for")
		if forAttr != "" {
			input = sc.page.Locator(fmt.Sprintf("#%s", forAttr))
			if inputCount, _ := input.Count(); inputCount > 0 {
				if err := input.First().Fill(value); err == nil {
					debugPrintf("✓ Filled '%s' via label-for '%s': %s\n", fieldLabel, forAttr, value)
					return nil
				}
			}
		}
	}

	// Strategy 4: Find input by placeholder
	input = sc.page.Locator(fmt.Sprintf("input[placeholder*='%s' i]", fieldLabel))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via placeholder: %s\n", fieldLabel, value)
			return nil
		}
	}

	// Strategy 5: Use Playwright's getByLabel
	input = sc.page.GetByLabel(fieldLabel)
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via getByLabel: %s\n", fieldLabel, value)
			return nil
		}
	}

	return fmt.Errorf("could not find input field '%s'", fieldLabel)
}

// selectBankOption opens the bank dropdown (Headless UI Listbox) and selects the given bank.
func (sc *E2EContext) selectBankOption(bankName string) error {
	debugPrintf("\n🏦 Selecting bank: %s\n", bankName)

	// Click the listbox button to open the dropdown
	// The bank select button has id="bank"
	bankButton := sc.page.Locator("#bank, button[id='bank']")
	if count, _ := bankButton.Count(); count == 0 {
		// Try alternative selectors
		bankButton = sc.page.Locator("button:near(:text('Bank'), 200)")
	}

	err := bankButton.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to open bank dropdown: %w", err)
	}

	// Wait for the dropdown to appear
	time.Sleep(300 * time.Millisecond)

	_ = sc.iTakeAScreenshot("bank-dropdown-open")

	// Select the bank option by role and text
	option := sc.page.GetByRole("option", playwright.PageGetByRoleOptions{
		Name: bankName,
	})
	if count, _ := option.Count(); count > 0 {
		err = option.First().Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		})
		if err == nil {
			debugPrintf("✓ Selected bank: %s\n", bankName)
			time.Sleep(300 * time.Millisecond)
			return nil
		}
	}

	// Fallback: try using text selector within the listbox
	listbox := sc.page.Locator("[role='listbox']")
	if count, _ := listbox.Count(); count > 0 {
		bankOption := listbox.Locator(fmt.Sprintf("li:has-text('%s'), [role='option']:has-text('%s')", bankName, bankName))
		if optCount, _ := bankOption.Count(); optCount > 0 {
			err = bankOption.First().Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(5000),
			})
			if err == nil {
				debugPrintf("✓ Selected bank via listbox: %s\n", bankName)
				time.Sleep(300 * time.Millisecond)
				return nil
			}
		}
	}

	return fmt.Errorf("could not select bank option '%s'", bankName)
}

// iShouldBeNavigatedToDashboard verifies navigation to a specific dashboard page.
func (sc *E2EContext) iShouldBeNavigatedToDashboard(dashboardName string) error {
	debugPrintf("\n🔍 Verifying navigation to dashboard: %s\n", dashboardName)

	// Wait for navigation to complete
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(1 * time.Second)

	_ = sc.iTakeAScreenshot(fmt.Sprintf("dashboard-%s", strings.ToLower(dashboardName)))

	currentURL := sc.page.URL()

	// "Home" is the root path /
	if strings.ToLower(dashboardName) == "home" {
		// Check URL is the root path (baseURL + "/" or just baseURL)
		trimmed := strings.TrimSuffix(currentURL, "/")
		base := strings.TrimSuffix(sc.baseURL, "/")
		if trimmed == base {
			debugPrintf("✓ On dashboard Home: %s\n", currentURL)
			return nil
		}
	}

	expectedPath := "/" + strings.ToLower(dashboardName)

	// Check URL contains the expected path
	if strings.Contains(strings.ToLower(currentURL), expectedPath) {
		debugPrintf("✓ On dashboard %s: %s\n", dashboardName, currentURL)
		return nil
	}

	// Also check page content for the dashboard name
	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, dashboardName) {
		debugPrintf("✓ Dashboard %s content found at: %s\n", dashboardName, currentURL)
		return nil
	}

	return fmt.Errorf("expected to be on %s dashboard, but at: %s", dashboardName, currentURL)
}

// theLinkedAccountShouldBeShownAs verifies that a linked account with the given
// display text (masked number or nickname) is visible on the /accounts list page.
func (sc *E2EContext) theLinkedAccountShouldBeShownAs(displayText string) error {
	debugPrintf("\n🔍 Verifying linked account shown as: %s\n", displayText)

	// Navigate to /accounts if not already there
	if !strings.HasSuffix(sc.page.URL(), "/accounts") {
		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(15000),
		})
	}

	// Retry loop: the account might take a moment to appear after creation
	for attempt := 0; attempt < 10; attempt++ {
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		})
		time.Sleep(1 * time.Second)

		_ = sc.iTakeAScreenshot(fmt.Sprintf("accounts-list-check-%s", strings.ReplaceAll(strings.ToLower(displayText), "*", "x")))

		// Look for account links containing the display text
		accountLinks := sc.page.Locator(fmt.Sprintf("a[href*='/accounts/']:has-text('%s')", displayText))
		if count, _ := accountLinks.Count(); count > 0 {
			debugPrintf("✓ Found linked account shown as: %s\n", displayText)
			return nil
		}

		// Broader check: look in the full page text
		allText, _ := sc.page.TextContent("body")
		if strings.Contains(allText, displayText) {
			debugPrintf("✓ Found '%s' in page text\n", displayText)
			return nil
		}

		debugPrintf("   ⚠️  '%s' not found on accounts page (attempt %d)\n", displayText, attempt+1)

		// Reload and retry
		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(10000),
		})
	}

	return fmt.Errorf("linked account with display text '%s' not found after 10 attempts", displayText)
}

// theLabelShouldBeShownForTheAccount navigates into the first bank account's
// detail page from /accounts and verifies a label/chip is visible (e.g., "Receive only").
func (sc *E2EContext) theLabelShouldBeShownForTheAccount(label string) error {
	debugPrintf("\n🏷️  Verifying label: %s\n", label)

	// Navigate to /accounts if not already there
	if !strings.HasSuffix(sc.page.URL(), "/accounts") && !strings.Contains(sc.page.URL(), "/accounts/") {
		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(15000),
		})
	}

	// If we're on /accounts (list), click into the first bank account
	if strings.HasSuffix(sc.page.URL(), "/accounts") {
		bankLink := sc.page.Locator("a[href*='/accounts/']").First()
		if err := bankLink.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("failed to click into account detail: %w", err)
		}
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		})
		time.Sleep(500 * time.Millisecond)
	}

	_ = sc.iTakeAScreenshot(fmt.Sprintf("label-%s", strings.ReplaceAll(strings.ToLower(label), " ", "-")))

	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, label) {
		debugPrintf("✓ Found label: %s\n", label)
		return nil
	}

	return fmt.Errorf("label '%s' not found on account detail page", label)
}

// iGiveTheLinkedAccountTheNickname sets a nickname for the linked account.
// Expects to be on or navigable to the account detail page.
func (sc *E2EContext) iGiveTheLinkedAccountTheNickname(nickname string) error {
	debugPrintf("\n✏️  Setting linked account nickname to: %s\n", nickname)

	// We should be on the account detail page (from the label check step).
	// Click the "Bank nickname" link to navigate to the edit page.
	nicknameLink := sc.page.Locator("a:has-text('nickname')")
	if count, _ := nicknameLink.Count(); count == 0 {
		// Fallback: look for a link to the /name edit page
		nicknameLink = sc.page.Locator("a[href$='/name']")
	}

	err := nicknameLink.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to click nickname link: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(500 * time.Millisecond)

	_ = sc.iTakeAScreenshot("nickname-edit-page")

	// Fill the nickname input (input name="name")
	nameInput := sc.page.Locator("input[name='name']")
	if count, _ := nameInput.Count(); count == 0 {
		nameInput = sc.page.Locator("input#name")
	}

	err = nameInput.First().Fill(nickname)
	if err != nil {
		return fmt.Errorf("failed to fill nickname input: %w", err)
	}

	_ = sc.iTakeAScreenshot("nickname-filled")

	// Click Save button
	saveBtn := sc.page.Locator("button:has-text('Save')")
	err = saveBtn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to click Save: %w", err)
	}

	// Wait for redirect back to detail page
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(1 * time.Second)

	debugPrintf("   📍 Current URL after save: %s\n", sc.page.URL())
	debugPrintf("✓ Set nickname to: %s\n", nickname)
	return nil
}

// toCamelCase converts a space-separated string to camelCase.
// e.g., "Account number" → "accountNumber"
func toCamelCase(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		if len(w) > 0 {
			result += strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return result
}
