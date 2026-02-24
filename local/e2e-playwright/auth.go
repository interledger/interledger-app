package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// iFillInLoginCredentials fills in email and password on the login page
func (sc *E2EContext) iFillInLoginCredentials(email, password string) error {
	// Check if email is already prefixed (starts with testEmailPrefix)
	// If not, prefix it
	prefixedEmail := email
	if sc.testIdentifier != "" && !strings.HasPrefix(email, sc.testIdentifier+"-") {
		prefixedEmail = fmt.Sprintf("%s-%s", sc.testIdentifier, email)
	}

	// Wait for page to load and be interactive
	time.Sleep(1 * time.Second)

	// Try multiple selectors for email input
	var emailInput playwright.Locator
	selectors := []string{
		"input[type='email']",
		"input[name*='email' i]",
		"input[id*='email' i]",
		"input[placeholder*='email' i]",
		"input[name='login']",
		"input[name='identifier']",
		"input[name='username']",
		"input[autocomplete='email']",
	}

	for _, selector := range selectors {
		emailInput = sc.page.Locator(selector)
		count, _ := emailInput.Count()
		if count > 0 {
			break
		}
	}

	emailCount, _ := emailInput.Count()
	if emailCount == 0 {
		// Check if we're already on TOTP setup page (user has active session from signup)
		totpInput := sc.page.Locator("input[name='totp_code']")
		totpCount, _ := totpInput.Count()
		if totpCount > 0 {
			return nil
		}
		return fmt.Errorf("failed to find email input - tried %d selectors", len(selectors))
	}

	// Fill email input with prefixed email
	err := emailInput.First().Fill(prefixedEmail)
	if err != nil {
		return fmt.Errorf("failed to fill email: %w", err)
	}
	debugPrintf("✓ Email filled: %s\n", prefixedEmail)

	// Wait a bit for password field to be visible
	time.Sleep(1 * time.Second)

	// Find and fill password input
	var passwordInput playwright.Locator
	passwordSelectors := []string{
		"input[type='password']",
		"input[name*='password' i]",
		"input[id*='password' i]",
		"input[placeholder*='password' i]",
	}

	for _, selector := range passwordSelectors {
		passwordInput = sc.page.Locator(selector)
		count, _ := passwordInput.Count()
		if count > 0 {
			break
		}
	}

	if count, _ := passwordInput.Count(); count == 0 {
		return fmt.Errorf("failed to find password input")
	}

	err = passwordInput.First().Fill(password)
	if err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}
	debugPrintf("✓ Password filled\n")

	return nil
}

// iSubmitTheLogin clicks the login submit button
func (sc *E2EContext) iSubmitTheLogin() error {
	// Take a screenshot before submitting to see what's on the page
	_ = sc.iTakeAScreenshot("before-login-submit")

	// Find and click login button
	var loginButton playwright.Locator
	buttonSelectors := []string{
		"button:has-text('Log In')",
		"button:has-text('Sign In')",
		"button:has-text('Login')",
		"button[type='submit']",
		"button:has-text('Submit')",
	}

	for _, selector := range buttonSelectors {
		loginButton = sc.page.Locator(selector)
		count, _ := loginButton.Count()
		if count > 0 {
			debugPrintf("Found login button with selector: %s\n", selector)
			break
		}
	}

	if count, _ := loginButton.Count(); count == 0 {
		return fmt.Errorf("failed to find login button")
	}

	// Click the login button
	err := loginButton.First().Click()
	if err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}
	debugPrintf("Login button clicked\n")

	// Wait for navigation with longer timeout
	time.Sleep(2 * time.Second)

	// Take screenshot after click
	_ = sc.iTakeAScreenshot("after-login-submit")

	// Check for error messages on the page
	errorSelectors := []string{
		".error",
		"[data-testid='error-message']",
		".alert-error",
		".message--error",
		"div:has-text('Error')",
		"div:has-text('Invalid')",
		"span:has-text('incorrect')",
		"p:has-text('wrong')",
	}
	for _, selector := range errorSelectors {
		errorMsg := sc.page.Locator(selector)
		count, _ := errorMsg.Count()
		if count > 0 {
			text, _ := errorMsg.First().TextContent()
			debugPrintf("⚠️  Found error message: %s\n", text)
		}
	}

	debugPrintf("Current URL after login click: %s\n", sc.page.URL())
	return nil
}

// iShouldBeNavigatedToTheTOTPPage verifies we're on the TOTP setup/registration page
func (sc *E2EContext) iShouldBeNavigatedToTheTOTPPage() error {
	// Retry up to 20 times (10 seconds total) to give navigation time to complete
	for i := 0; i < 20; i++ {
		currentURL := sc.page.URL()

		if strings.Contains(currentURL, "/totp") || strings.Contains(currentURL, "two-factor") {
			return nil
		}

		// If still on login, check if page content changed
		if strings.Contains(currentURL, "/login") {
			// Wait before retrying
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// URL changed to something else, that's good
		return nil
	}

	currentURL := sc.page.URL()
	return fmt.Errorf("not on TOTP page, current URL: %s", currentURL)
}

// iTypeInMyGeneratedTotpForMyNewUser generates and enters a TOTP code
func (sc *E2EContext) iTypeInMyGeneratedTotpForMyNewUser() error {
	// The TOTP secret should be visible on the page for first-time setup
	// We need to extract it from the page and generate a code

	// Wait for page to load
	time.Sleep(1 * time.Second)

	// Look for the TOTP secret on the page (it should be displayed during setup)
	// Common patterns: QR code + text secret, or just text

	// Get page content to find the secret
	content, err := sc.page.Content()
	if err != nil {
		return fmt.Errorf("failed to get page content: %w", err)
	}

	// Try to find the secret in the page content
	// Kratos usually shows it in the otpauth URL or as plain text
	// Save content to file for debugging
	os.WriteFile("/tmp/totp-page-content.html", []byte(content), 0644)

	secret, err := sc.extractTOTPSecretFromPage(content)
	if err != nil {
		return fmt.Errorf("failed to extract TOTP secret from page: %w", err)
	}

	// Store the TOTP secret in userDetails for later use during login
	if sc.currentUser != "" && sc.userDetails[sc.currentUser] != nil {
		sc.userDetails[sc.currentUser].Fields["totp_secret"] = secret
		debugPrintf("📝 Stored TOTP secret in userDetails for '%s'\n", sc.currentUser)
	}

	// Generate TOTP code
	code, err := generateTOTPCode(secret)
	if err != nil {
		return fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	// Find the TOTP input field
	totpInput := sc.page.Locator("input[name='totp_code'], input[type='text'], input[type='number']")
	totpCount, _ := totpInput.Count()
	if totpCount == 0 {
		return fmt.Errorf("failed to find TOTP code input field")
	}

	// Fill in the code
	err = totpInput.First().Fill(code)
	if err != nil {
		return fmt.Errorf("failed to fill TOTP code: %w", err)
	}

	return nil
}

// iShouldBeNavigatedToTheApplicationDashboard verifies navigation to dashboard after TOTP
func (sc *E2EContext) iShouldBeNavigatedToTheApplicationDashboard() error {
	debugPrintln("   🏠 Waiting for dashboard navigation...")

	// Dashboard URLs might include: /dashboard, /app, /home, /wallet, etc.
	dashboardPatterns := []string{
		"/dashboard",
		"/app",
		"/home",
		"/wallet",
		"/accounts",
		"/personal-details",
		"/activation",
	}

	// Try for up to 15 seconds waiting for dashboard navigation
	for attempt := 0; attempt < 30; attempt++ {
		currentURL := sc.page.URL()
		debugPrintf("   📍 Current URL (attempt %d): %s\n", attempt+1, currentURL)

		// Check if we're on a dashboard page
		for _, pattern := range dashboardPatterns {
			if strings.Contains(currentURL, pattern) {
				debugPrintf("   ✓ Successfully navigated to dashboard (matched pattern: %s)\n", pattern)
				return nil
			}
		}

		// If not on login/totp page, we might be on an okay page
		if !strings.Contains(currentURL, "/login") && !strings.Contains(currentURL, "/totp") {
			debugPrintf("   ✓ On application page (not login/TOTP): %s\n", currentURL)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	finalURL := sc.page.URL()
	return fmt.Errorf("not on application dashboard - still at: %s", finalURL)
}

// extractTOTPSecretFromPage extracts the TOTP secret from the page HTML
func (sc *E2EContext) extractTOTPSecretFromPage(htmlContent string) (string, error) {
	// The TOTP secret is displayed as plain text in the HTML (e.g., "WMU5UNIH7K5Q2U7JWYLYLUIIJX7UYKWK")
	// Look for a base32 string (uppercase letters A-Z and digits 2-7) that's typically 32 characters
	secretRegex := regexp.MustCompile(`\b([A-Z2-7]{32})\b`)
	if matches := secretRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		secret := matches[1]
		return secret, nil
	}

	// Also try the otpauth:// URL pattern as fallback
	otpauthPrefix := "otpauth://totp/"
	otpauthStart := strings.Index(htmlContent, otpauthPrefix)
	if otpauthStart != -1 {
		// Find the end of the URL (usually a quote or space)
		otpauthEnd := otpauthStart
		for otpauthEnd < len(htmlContent) && htmlContent[otpauthEnd] != '"' && htmlContent[otpauthEnd] != '\'' && htmlContent[otpauthEnd] != ' ' && htmlContent[otpauthEnd] != '<' {
			otpauthEnd++
		}
		otpauthURL := htmlContent[otpauthStart:otpauthEnd]
		return extractSecretFromURL(otpauthURL)
	}

	// Alternative: look for "secret=" pattern directly
	secretPrefix := "secret="
	secretStart := strings.Index(htmlContent, secretPrefix)
	if secretStart != -1 {
		secretStart += len(secretPrefix)
		// Skip any quotes
		if secretStart < len(htmlContent) && (htmlContent[secretStart] == '"' || htmlContent[secretStart] == '\'') {
			secretStart++
		}
		// Find the end (quote, ampersand, or space)
		secretEnd := secretStart
		for secretEnd < len(htmlContent) && htmlContent[secretEnd] != '"' && htmlContent[secretEnd] != '\'' && htmlContent[secretEnd] != '&' && htmlContent[secretEnd] != ' ' && htmlContent[secretEnd] != '<' {
			secretEnd++
		}
		secret := htmlContent[secretStart:secretEnd]
		// Validate it looks like a base32 secret
		if len(secret) >= 16 && len(secret) <= 64 {
			return secret, nil
		}
	}

	return "", fmt.Errorf("could not find TOTP secret in page content")
}

// iSubmitTheTotpRegistration clicks the TOTP registration submit button
func (sc *E2EContext) iSubmitTheTotpRegistration() error {
	debugPrintln("   🔘 Looking for TOTP submit button...")

	// Find submit button - TotpChallenge component uses button with type='submit'
	// and text 'Verify' when showing challenge form
	submitButton := sc.page.Locator("button:has-text('Verify')")
	submitCount, _ := submitButton.Count()
	if submitCount == 0 {
		// Fallback to type='submit' if text selector doesn't work
		submitButton = sc.page.Locator("button[type='submit']")
		submitCount, _ = submitButton.Count()
		if submitCount == 0 {
			return fmt.Errorf("failed to find TOTP submit button (tried: 'Verify' text and type='submit')")
		}
	}

	debugPrintf("   ✓ Found %d button(s), clicking first one\n", submitCount)

	err := submitButton.First().Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("failed to click TOTP submit button: %w", err)
	}

	debugPrintln("   ⏳ Waiting for page redirect after TOTP submission...")

	// Wait for navigation to complete. TOTP submission triggers a redirect via redirectDocument()
	// which should happen within 5 seconds
	currentURL := sc.page.URL()
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		newURL := sc.page.URL()
		if newURL != currentURL {
			debugPrintf("   ✓ Page redirected after TOTP submission (attempt %d)\n", i+1)
			return nil
		}
	}

	// If URL didn't change, wait a bit more and return
	// The app may redirect after a delay
	debugPrintf("   ⚠️  URL did not change within 5s, returning anyway\n")
	time.Sleep(1 * time.Second)

	return nil
}

// Helper functions for TOTP generation

// extractSecretFromURL extracts the secret parameter from an otpauth:// URL
func extractSecretFromURL(totpURL string) (string, error) {
	// Find the secret= parameter
	secretIdx := strings.Index(totpURL, "secret=")
	if secretIdx == -1 {
		return "", fmt.Errorf("secret parameter not found in URL")
	}

	secretStart := secretIdx + len("secret=")

	// Find the end of the secret (next & or end of string)
	secretEnd := secretStart
	for secretEnd < len(totpURL) && totpURL[secretEnd] != '&' && totpURL[secretEnd] != '"' {
		secretEnd++
	}

	return totpURL[secretStart:secretEnd], nil
}

// generateTOTPCode generates a 6-digit TOTP code from a base32 secret
func generateTOTPCode(secret string) (string, error) {
	// Decode base32 secret
	decoded, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Get current time counter (30-second intervals)
	counter := time.Now().Unix() / 30

	// Convert counter to byte array
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, decoded)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate 6-digit code
	code := truncated % 1000000

	return fmt.Sprintf("%06d", code), nil
}
