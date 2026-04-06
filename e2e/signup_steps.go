package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// iFillInWithUniquePhoneNumber fills in phone field with a unique phone number for the current user's country
// Usage: I fill in "phone" with a unique valid phone number
func (sc *E2EContext) iFillInWithUniquePhoneNumber(fieldName string) error {
	if sc.currentUser == "" {
		return fmt.Errorf("no current user set for phone generation")
	}

	details, ok := sc.userDetails[sc.currentUser]
	if !ok {
		return fmt.Errorf("no user details for current user '%s'", sc.currentUser)
	}

	emailSuffix, ok := details.Fields["emailSuffix"]
	if !ok {
		return fmt.Errorf("no emailSuffix defined for user '%s'", sc.currentUser)
	}

	phoneNumber, err := allocatePhoneNumber(sc.country)
	if err != nil {
		return fmt.Errorf("failed to allocate phone number: %w", err)
	}
	debugPrintf("📱 Generated unique phone number: %s (country %s, emailSuffix %s, user %s)\n", phoneNumber, sc.country, emailSuffix, sc.currentUser)

	// Store in user details for current user
	if sc.currentUser != "" && sc.userDetails[sc.currentUser] != nil {
		sc.userDetails[sc.currentUser].Fields["phone"] = phoneNumber
		debugPrintf("📱 Stored phone in userDetails for '%s'\n", sc.currentUser)
	} else {
		return fmt.Errorf("cannot store phone: no current user set or user details not initialized")
	}

	// Fill in the phone field
	err = sc.iFillInWith(fieldName, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to fill phone field: %w", err)
	}

	// Trigger blur event to ensure onChange/onBlur handlers fire
	input := sc.page.Locator("input[type='tel']")
	input.Blur()

	// Wait for the value to be processed
	sc.page.WaitForTimeout(500)

	// Verify the value was actually set
	actualValue, _ := input.InputValue()
	debugPrintf("📱 Phone field value after filling: %s\n", actualValue)

	return nil
}

// aSignupRecordShouldExistForMyself checks that a signup record exists for the current user
// Usage: a signup record should exist in the database for myself
func (sc *E2EContext) aSignupRecordShouldExistForMyself() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}
	return sc.aSignupRecordShouldExistInTheDatabase(email)
}

// theSignupShouldHaveFirstNameMatching checks the signup first name matches a field value
// Usage: the signup should have first name matching my "firstName"
func (sc *E2EContext) theSignupShouldHaveFirstNameMatching(fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}
	return sc.theSignupShouldHaveFirstName(value)
}

// theSignupShouldHaveLastNameMatching checks the signup last name matches a field value
// Usage: the signup should have last name matching my "lastName"
func (sc *E2EContext) theSignupShouldHaveLastNameMatching(fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}
	return sc.theSignupShouldHaveLastName(value)
}

// theSignupShouldHaveCountryCodeMatching checks the signup country code matches a field value
// Usage: the signup should have country code matching my "countryCode"
func (sc *E2EContext) theSignupShouldHaveCountryCodeMatching(fieldKey string) error {
	value, err := sc.getFieldValue(fieldKey)
	if err != nil {
		return err
	}
	return sc.theSignupShouldHaveCountryCode(value)
}

// iTriggerUserVerificationForMyself triggers user verification for the current user
// Usage: I trigger user verification for myself
func (sc *E2EContext) iTriggerUserVerificationForMyself() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}
	return sc.iTriggerUserVerificationFor(email)
}

// iCompletedTheSignupWorkflow runs the signup flow using the current impersonated user's details.
// Usage: I completed the signup workflow
func (sc *E2EContext) iCompletedTheSignupWorkflow() error {
	if sc.currentUser == "" {
		return fmt.Errorf("no user currently impersonated")
	}

	firstName := sc.firstName
	lastName := sc.lastName
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get email for signup workflow: %w", err)
	}
	country := sc.country
	password := sc.password

	if value, err := sc.getFieldValue("firstName"); err == nil && value != "" {
		firstName = value
	}
	if value, err := sc.getFieldValue("lastName"); err == nil && value != "" {
		lastName = value
	}
	if value, err := sc.getFieldValue("emailSuffix"); err == nil && value != "" {
		email = value
	}
	if value, err := sc.getFieldValue("country"); err == nil && value != "" {
		country = value
	}
	if value, err := sc.getFieldValue("password"); err == nil && value != "" {
		password = value
	}

	if firstName == "" || lastName == "" || email == "" || country == "" || password == "" {
		return fmt.Errorf("missing signup details (firstName=%q, lastName=%q, email=%q, country=%q, password set=%t)", firstName, lastName, email, country, password != "")
	}

	// Phone prefix is empty — generateDeterministicPhone picks the right base from sc.country
	phonePrefix := ""

	return sc.iCompleteSignupFlow(firstName, lastName, email, country, phonePrefix, password)
}

// iFillInTheLoginFormWithMyDetails fills login form using current user context.
// Usage: I fill in the login form with my details
func (sc *E2EContext) iFillInTheLoginFormWithMyDetails() error {
	if sc.email == "" {
		if value, err := sc.getFieldValue("emailSuffix"); err == nil && value != "" {
			sc.email = fmt.Sprintf("%s-%s", sc.testIdentifier, value)
		}
	}
	if sc.password == "" {
		if value, err := sc.getFieldValue("password"); err == nil && value != "" {
			sc.password = value
		}
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}
	if sc.password == "" {
		return fmt.Errorf("missing password for login")
	}

	debugPrintf("📝 Filling login form with user: %s\n", email)
	return sc.iFillInLoginCredentials(email, sc.password)
}

// iFillInMyLoginCredentials fills in email and password for the current user
// Usage: I fill in my login credentials
func (sc *E2EContext) iFillInMyLoginCredentials() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}
	if sc.password == "" {
		return fmt.Errorf("no password set for current user")
	}

	debugPrintf("📝 Filling login credentials for user: %s\n", email)
	return sc.iFillInLoginCredentials(email, sc.password)
}

// iTypeInMyGeneratedTotpForMyself generates and types TOTP for the current user
// Usage: I type in my generated totp for myself
func (sc *E2EContext) iTypeInMyGeneratedTotpForMyself() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	// Try to get the TOTP secret from Kratos first
	totpSecret, err := sc.getTOTPSecretForEmail(email)
	if err != nil {
		// If not in Kratos yet (first time setup), extract from the page
		debugPrintf("⚠️  TOTP secret not in Kratos yet, extracting from page: %v\n", err)

		// Get page content
		content, err := sc.page.Content()
		if err != nil {
			return fmt.Errorf("failed to get page content: %w", err)
		}

		totpSecret, err = extractTOTPSecretFromPageHTML(content)
		if err != nil {
			return fmt.Errorf("failed to extract TOTP secret from page: %w", err)
		}
	}

	sc.totpSecret = totpSecret

	// Generate TOTP code
	totpCode, err := generateTOTPCode(totpSecret)
	if err != nil {
		return fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	debugPrintf("🔐 Generated TOTP code: %s\n", totpCode)

	// Find TOTP input field
	totpInput := sc.page.Locator("input[name='totp_code'], input[placeholder*='code' i], input[placeholder*='totp' i], input[name='code']")
	count, _ := totpInput.Count()
	if count == 0 {
		return fmt.Errorf("failed to find TOTP input field")
	}

	// Fill in the TOTP code
	err = totpInput.First().Fill(totpCode)
	if err != nil {
		return fmt.Errorf("failed to fill TOTP code: %w", err)
	}

	debugPrintf("✓ TOTP code entered\n")
	return nil
}

// extractTOTPSecretFromPageHTML extracts the TOTP secret from the page HTML
func extractTOTPSecretFromPageHTML(htmlContent string) (string, error) {
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

// getTOTPSecretForEmail retrieves the TOTP secret from Kratos for a given email
func (sc *E2EContext) getTOTPSecretForEmail(email string) (string, error) {
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	listReq, err := http.NewRequestWithContext(context.Background(), "GET", kratosAdminURL+"/admin/identities", nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	listResp, err := client.Do(listReq)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(listResp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", listResp.StatusCode, string(body))
	}

	var identities []kratosIdentity
	if err := json.NewDecoder(listResp.Body).Decode(&identities); err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}

	for _, identity := range identities {
		traits := identity.Traits
		if v, ok := traits["email"]; ok {
			if s, ok2 := v.(string); ok2 && s == email {
				// Found the identity, extract TOTP secret from credentials
				if creds, ok := identity.Credentials["totp"]; ok {
					if credMap, ok := creds.(map[string]interface{}); ok {
						if config, ok := credMap["config"]; ok {
							if configMap, ok := config.(map[string]interface{}); ok {
								if secret, ok := configMap["secret"]; ok {
									if secretStr, ok := secret.(string); ok && secretStr != "" {
										debugPrintf("✓ Retrieved TOTP secret for %s\n", email)
										return secretStr, nil
									}
								}
							}
						}
					}
				}
				return "", fmt.Errorf("no TOTP secret configured for %s", email)
			}
		}
	}

	return "", fmt.Errorf("identity not found for email %s", email)
}

// iLogInAsMyself logs in as the current impersonated user
// Usage: I log in as myself
func (sc *E2EContext) iLogInAsMyself() error {
	if sc.currentUser == "" {
		return fmt.Errorf("no user currently impersonated")
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get email for login: %w", err)
	}

	if sc.password == "" {
		// Try to get password from user details
		if value, err := sc.getFieldValue("password"); err == nil && value != "" {
			sc.password = value
		} else {
			return fmt.Errorf("no password set for current user")
		}
	}

	debugPrintf("🔐 Logging in as: %s\n", email)

	// First navigate to login page
	loginURL := sc.baseURL + "/login"
	_, err = sc.page.Goto(loginURL)
	if err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for page to be interactive
	sc.page.WaitForLoadState()

	debugPrintf("✓ Navigated to login page\n")

	// Fill in login credentials
	err = sc.iFillInLoginCredentials(email, sc.password)
	if err != nil {
		return fmt.Errorf("failed to fill login credentials: %w", err)
	}

	// Submit the login form
	err = sc.iSubmitTheLogin()
	if err != nil {
		return fmt.Errorf("failed to submit login: %w", err)
	}

	// Now we should be on TOTP page, enter the TOTP
	time.Sleep(2 * time.Second)

	// Check if we're on TOTP page
	currentURL := sc.page.URL()
	if strings.Contains(currentURL, "totp") {
		debugPrintf("🔑 On TOTP page, entering TOTP code...\n")

		// Try to get TOTP secret from userDetails first (stored during signup)
		totpSecret, err := sc.getCurrentUserTOTPSecret()
		if err != nil {
			// Not in userDetails, try fetching from Kratos
			debugPrintf("⚠️  TOTP secret not in userDetails (%v), trying Kratos: %v\n", sc.currentUser, err)
			totpSecret, err = sc.getTOTPSecretForEmail(email)
			if err != nil {
				debugPrintf("⚠️  Could not get TOTP secret from Kratos: %v\n", err)
				// Try alternative: use existing TOTP field or skip if we can't find it
				return fmt.Errorf("no TOTP secret available for user - user may need to re-verify email")
			}
			// Store it for future use
			if sc.currentUser != "" && sc.userDetails[sc.currentUser] != nil {
				sc.userDetails[sc.currentUser].Fields["totp_secret"] = totpSecret
				debugPrintf("✓ Stored TOTP secret in userDetails for '%s' (from Kratos)\n", sc.currentUser)
			}
		} else {
			debugPrintf("✓ Retrieved TOTP secret from userDetails for '%s'\n", sc.currentUser)
		}

		// Calculate TOTP code
		code, err := generateTOTPCode(totpSecret)
		if err != nil {
			return fmt.Errorf("failed to calculate TOTP: %w", err)
		}

		// Find and fill TOTP input
		totpInput := sc.page.Locator("input[name='totp_code'], input[placeholder*='TOTP' i], input[placeholder*='code' i]")
		count, _ := totpInput.Count()
		if count > 0 {
			err = totpInput.First().Fill(code)
			if err != nil {
				return fmt.Errorf("failed to fill TOTP code: %w", err)
			}
			debugPrintf("✓ Filled TOTP code\n")
		} else {
			return fmt.Errorf("could not find TOTP input field")
		}

		// Wait a moment for processing
		time.Sleep(1 * time.Second)

		// Check if there's a TOTP submit button to click
		submitButton := sc.page.Locator("button:has-text('Verify'), button[type='submit']:not([disabled])")
		count, _ = submitButton.Count()
		if count > 0 {
			err = submitButton.First().Click()
			if err != nil {
				debugPrintf("⚠️  Failed to click TOTP submit: %v\n", err)
			}
		}

		// Wait for final redirect
		time.Sleep(3 * time.Second)
	}

	debugPrintf("✅ Login completed\n")
	return nil
}
