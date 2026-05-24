package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iNavigateToTheDashboardPage navigates to a named dashboard page.
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
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	})

	debugPrintf("✓ Navigated to dashboard %s: %s\n", page, url)
	return nil
}

// iPressOn clicks on any clickable element (link or button) with the given text.
func (sc *E2EContext) iPressOn(text string) error {
	debugPrintf("\n👆 Pressing on: %s\n", text)

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

		for j := 0; j < count; j++ {
			el := allMatches.Nth(j)
			visible, _ := el.IsVisible()
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
				debugPrintf("✓ Pressed on: %s\n", text)
				return nil
			}
		}
	}

	return fmt.Errorf("could not find clickable element with text '%s'", text)
}

// iShouldSeeTheForm verifies a form with the given title is visible on the page.
func (sc *E2EContext) iShouldSeeTheForm(formTitle string) error {
	debugPrintf("\n📋 Verifying form: %s\n", formTitle)

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(500 * time.Millisecond)

	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, formTitle) {
		debugPrintf("✓ Found form: %s\n", formTitle)
		return nil
	}

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
func (sc *E2EContext) iFillInFieldWith(fieldLabel, value string) error {
	debugPrintf("\n📝 Filling field '%s' with '%s'\n", fieldLabel, value)

	camelID := toCamelCase(fieldLabel)

	input := sc.page.Locator(fmt.Sprintf("input#%s", camelID))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via id '%s'\n", fieldLabel, camelID)
			return nil
		}
	}

	input = sc.page.Locator(fmt.Sprintf("input[name='%s']", camelID))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via name '%s'\n", fieldLabel, camelID)
			return nil
		}
	}

	label := sc.page.Locator(fmt.Sprintf("label:has-text('%s')", fieldLabel))
	if count, _ := label.Count(); count > 0 {
		forAttr, _ := label.First().GetAttribute("for")
		if forAttr != "" {
			input = sc.page.Locator(fmt.Sprintf("#%s", forAttr))
			if inputCount, _ := input.Count(); inputCount > 0 {
				if err := input.First().Fill(value); err == nil {
					debugPrintf("✓ Filled '%s' via label-for '%s'\n", fieldLabel, forAttr)
					return nil
				}
			}
		}
	}

	input = sc.page.Locator(fmt.Sprintf("input[placeholder*='%s' i]", fieldLabel))
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via placeholder\n", fieldLabel)
			return nil
		}
	}

	input = sc.page.GetByLabel(fieldLabel)
	if count, _ := input.Count(); count > 0 {
		if err := input.First().Fill(value); err == nil {
			debugPrintf("✓ Filled '%s' via getByLabel\n", fieldLabel)
			return nil
		}
	}

	return fmt.Errorf("could not find input field '%s'", fieldLabel)
}

// selectBankOption opens the bank dropdown and selects the given bank.
func (sc *E2EContext) selectBankOption(bankName string) error {
	debugPrintf("\n🏦 Selecting bank: %s\n", bankName)

	bankButton := sc.page.Locator("#bank, button[id='bank']")
	if count, _ := bankButton.Count(); count == 0 {
		bankButton = sc.page.Locator("button:near(:text('Bank'), 200)")
	}

	if err := bankButton.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to open bank dropdown: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	option := sc.page.GetByRole("option", playwright.PageGetByRoleOptions{
		Name: bankName,
	})
	if count, _ := option.Count(); count > 0 {
		if err := option.First().Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		}); err == nil {
			debugPrintf("✓ Selected bank: %s\n", bankName)
			time.Sleep(300 * time.Millisecond)
			return nil
		}
	}

	listbox := sc.page.Locator("[role='listbox']")
	if count, _ := listbox.Count(); count > 0 {
		bankOption := listbox.Locator(fmt.Sprintf("li:has-text('%s'), [role='option']:has-text('%s')", bankName, bankName))
		if optCount, _ := bankOption.Count(); optCount > 0 {
			if err := bankOption.First().Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(5000),
			}); err == nil {
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

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(1 * time.Second)

	currentURL := sc.page.URL()

	if strings.ToLower(dashboardName) == "home" {
		trimmed := strings.TrimSuffix(currentURL, "/")
		base := strings.TrimSuffix(sc.baseURL, "/")
		if trimmed == base {
			debugPrintf("✓ On dashboard Home: %s\n", currentURL)
			return nil
		}
	}

	expectedPath := "/" + strings.ToLower(dashboardName)
	if strings.Contains(strings.ToLower(currentURL), expectedPath) {
		debugPrintf("✓ On dashboard %s: %s\n", dashboardName, currentURL)
		return nil
	}

	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, dashboardName) {
		debugPrintf("✓ Dashboard %s content found at: %s\n", dashboardName, currentURL)
		return nil
	}

	return fmt.Errorf("expected to be on %s dashboard, but at: %s", dashboardName, currentURL)
}

// theLinkedAccountShouldBeShownAs verifies that a linked account is visible on /accounts.
func (sc *E2EContext) theLinkedAccountShouldBeShownAs(displayText string) error {
	debugPrintf("\n🔍 Verifying linked account shown as: %s\n", displayText)

	if !strings.HasSuffix(sc.page.URL(), "/accounts") {
		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(15000),
		})
	}

	for attempt := 0; attempt < 10; attempt++ {
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(10000),
		})
		time.Sleep(1 * time.Second)

		accountLinks := sc.page.Locator(fmt.Sprintf("a[href*='/accounts/']:has-text('%s')", displayText))
		if count, _ := accountLinks.Count(); count > 0 {
			debugPrintf("✓ Found linked account shown as: %s\n", displayText)
			return nil
		}

		allText, _ := sc.page.TextContent("body")
		if strings.Contains(allText, displayText) {
			debugPrintf("✓ Found '%s' in page text\n", displayText)
			return nil
		}

		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(10000),
		})
	}

	return fmt.Errorf("linked account with display text '%s' not found after 10 attempts", displayText)
}

// theLabelShouldBeShownForTheAccount verifies a label/chip is visible on the account detail page.
func (sc *E2EContext) theLabelShouldBeShownForTheAccount(label string) error {
	debugPrintf("\n🏷️  Verifying label: %s\n", label)

	if !strings.HasSuffix(sc.page.URL(), "/accounts") && !strings.Contains(sc.page.URL(), "/accounts/") {
		_, _ = sc.page.Goto(sc.baseURL+"/accounts", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(15000),
		})
	}

	if strings.HasSuffix(sc.page.URL(), "/accounts") {
		bankLink := sc.page.Locator("a[href*='/accounts/']").First()
		if err := bankLink.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("failed to click into account detail: %w", err)
		}
		sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(10000),
		})
		time.Sleep(500 * time.Millisecond)
	}

	allText, _ := sc.page.TextContent("body")
	if strings.Contains(allText, label) {
		debugPrintf("✓ Found label: %s\n", label)
		return nil
	}

	return fmt.Errorf("label '%s' not found on account detail page", label)
}

// iGiveTheLinkedAccountTheNickname sets a nickname for the linked account.
func (sc *E2EContext) iGiveTheLinkedAccountTheNickname(nickname string) error {
	debugPrintf("\n✏️  Setting linked account nickname to: %s\n", nickname)

	nicknameLink := sc.page.Locator("a:has-text('nickname')")
	if count, _ := nicknameLink.Count(); count == 0 {
		nicknameLink = sc.page.Locator("a[href$='/name']")
	}

	if err := nicknameLink.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click nickname link: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(500 * time.Millisecond)

	nameInput := sc.page.Locator("input[name='name']")
	if count, _ := nameInput.Count(); count == 0 {
		nameInput = sc.page.Locator("input#name")
	}

	if err := nameInput.First().Fill(nickname); err != nil {
		return fmt.Errorf("failed to fill nickname input: %w", err)
	}

	saveBtn := sc.page.Locator("button:has-text('Save')")
	if err := saveBtn.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("failed to click Save: %w", err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(10000),
	})
	time.Sleep(1 * time.Second)

	debugPrintf("✓ Set nickname to: %s\n", nickname)
	return nil
}

// toCamelCase converts a space-separated string to camelCase.
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
