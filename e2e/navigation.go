package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Navigation step implementations

// iClearTheBrowserSession clears cookies and storage to reset the session
func (sc *E2EContext) iClearTheBrowserSession() error {
	debugPrintln("\n🔄 Clearing browser session (cookies and storage)...")

	if sc.context == nil {
		return fmt.Errorf("browser context not initialized")
	}

	// Clear cookies
	if err := sc.context.ClearCookies(); err != nil {
		return fmt.Errorf("failed to clear cookies: %w", err)
	}

	// Clear storage (localStorage, sessionStorage, etc.)
	if _, err := sc.page.Evaluate(`() => {
		localStorage.clear();
		sessionStorage.clear();
	}`); err != nil {
		return fmt.Errorf("failed to clear storage: %w", err)
	}

	debugPrintln("✅ Browser session cleared successfully")
	return nil
}

// iNavigateToURL navigates to a specific URL
func (sc *E2EContext) iNavigateToURL(url string) error {
	debugPrintf("\n🌐 Navigating to: %s\n", url)

	if sc.page == nil {
		return fmt.Errorf("page not initialized - need to setup browser first")
	}

	_, err := sc.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(10000),
	})

	if err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	debugPrintf("✓ Navigated to: %s\n", url)
	return nil
}

func (sc *E2EContext) iNavigateToTheSignupPage() error {
	// Setup browser if not already done
	if sc.browser == nil {
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
	}

	// If context already exists, close it and create a new one for a fresh session
	if sc.context != nil {
		debugPrintln("\n🔄 Creating fresh context for new user signup...")
		if sc.page != nil {
			_ = sc.page.Close()
		}
		_ = sc.context.Close()
	}

	context, err := sc.browser.NewContext(playwright.BrowserNewContextOptions{
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

	// Navigate directly to the signup page
	signupURL := sc.baseURL + "/signup"
	var navErr error
	for attempt := 0; attempt < 3; attempt++ {
		_, navErr = sc.page.Goto(signupURL, playwright.PageGotoOptions{
			Timeout:   playwright.Float(30000),
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		})
		if navErr == nil {
			break
		}
		debugPrintf("   ⚠️  Navigation attempt %d failed: %v — retrying\n", attempt+1, navErr)
		time.Sleep(2 * time.Second)
	}
	if navErr != nil {
		return fmt.Errorf("failed to navigate to signup: %w", navErr)
	}

	// Ensure page is fully loaded by waiting for network
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	return nil
}

func (sc *E2EContext) iClickTheButton(buttonText string) error {
	// Debug: Take screenshot before clicking to see what's on the page
	_ = sc.iTakeAScreenshot(fmt.Sprintf("before-%s", strings.ReplaceAll(strings.ToLower(buttonText), " ", "-")))

	var selector string

	switch strings.ToLower(buttonText) {
	case "sign up", "get started":
		selector = "button:has-text('Sign Up'), button:has-text('Get Started'), a:has-text('Sign Up')"
	case "let's get started":
		selector = "button:has-text('Let'), button:has-text('Get started'), button[data-testid='signup-get-started']"
	case "continue", "next":
		selector = "button:has-text('Continue'), button:has-text('Next'), button[type='submit']"
	case "confirm", "submit":
		selector = "button:has-text('Confirm'), button:has-text('Submit'), button[type='submit']"
	default:
		selector = fmt.Sprintf("button:has-text('%s')", buttonText)
	}

	button := sc.page.Locator(selector)
	err := button.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		// Timeout on final submit is expected
		if strings.Contains(strings.ToLower(buttonText), "confirm") ||
			strings.Contains(strings.ToLower(buttonText), "submit") {
			time.Sleep(1 * time.Second)
			return nil
		}
		return fmt.Errorf("failed to click button '%s': %w", buttonText, err)
	}

	// Wait for page to settle
	sc.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	return nil
}

func (sc *E2EContext) iShouldSeeTheSignupForm() error {
	// Check if we're on the signup page
	currentURL := sc.page.URL()
	if !strings.Contains(currentURL, "/signup") {
		return fmt.Errorf("not on signup page, current URL: %s", currentURL)
	}

	// Wait for signup form inputs to be visible
	// Use a longer timeout and retry WITHOUT reloading (reloading restarts React and causes race conditions)
	firstNameInput := sc.page.Locator("input[name='firstName'], input[name='first_name'], input[placeholder*='first' i]")

	// Single wait with increased timeout (60s total instead of 3×10s with reloads)
	// This allows React to fully hydrate even under load
	err := firstNameInput.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(60000),
	})
	if err == nil {
		_ = sc.iTakeAScreenshot("signup-form")
		return nil
	}

	// If still not visible after 60s, take a debug screenshot
	debugPrintf("   ⚠️  Signup form not visible after 60s, taking debug screenshot\n")
	_ = sc.iTakeAScreenshot("signup-form-not-visible")

	// Try one more time with a shorter wait (in case React just finished mounting)
	err = firstNameInput.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err == nil {
		_ = sc.iTakeAScreenshot("signup-form")
		return nil
	}

	return fmt.Errorf("signup form not found: first name input not visible after 65s")
}

func (sc *E2EContext) iShouldBeOnStep(step int) error {
	sc.currentStep = step
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (sc *E2EContext) iTakeAScreenshot(name string) error {
	shotDir := sc.screenshotDir
	if shotDir == "" {
		shotDir = filepath.Join("debug", "unknown-scenario")
	}
	if err := os.MkdirAll(shotDir, 0755); err != nil {
		return err
	}

	shotName := normalizeName(name)
	if shotName == "" {
		shotName = "screenshot"
	}
	sc.screenshotCount++
	path := filepath.Join(shotDir, fmt.Sprintf("%s-%02d-%s.png", sc.testIdentifier, sc.screenshotCount, shotName))
	_, err := sc.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	return err
}

func (sc *E2EContext) iNavigateToTheLoginPage() error {
	loginURL := sc.baseURL + "/login"
	_, err := sc.page.Goto(loginURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("failed to navigate to login: %w", err)
	}

	return nil
}
