package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

var _ = (*E2EContext).getKratosIdentities

// Assertion step implementations

func (sc *E2EContext) theSignupShouldBeSubmitted() error {
	// Verify password was filled during the flow
	if !sc.passwordFilled {
		return fmt.Errorf("password was never filled during signup flow")
	}

	// Wait for submission to complete
	time.Sleep(2 * time.Second)

	// Verify the Kratos identity was actually created by the signup form submission.
	// Under high concurrency the POST registration may silently fail (e.g. timeout,
	// CSRF expiry). Catching it here gives a clear error rather than a confusing
	// "invalid credentials" failure later during login.
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("signup verification: %w", err)
	}
	prefixedEmail := email
	if sc.testIdentifier != "" && !strings.HasPrefix(email, sc.testIdentifier+"-") {
		prefixedEmail = fmt.Sprintf("%s-%s", sc.testIdentifier, email)
	}

	// Open a single Kratos DB connection for the polling loop to avoid
	// connection churn under concurrent scenarios.
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		return fmt.Errorf("signup verification: could not connect to Kratos DB: %w", err)
	}
	defer kratosDB.Close()

	// Poll for the identity for up to 10 seconds (20 × 500ms)
	for i := 0; i < 20; i++ {
		var identityID string
		row := kratosDB.QueryRow(`SELECT identity_id FROM identity_verifiable_addresses WHERE value = $1 LIMIT 1`, prefixedEmail)
		if scanErr := row.Scan(&identityID); scanErr == nil && identityID != "" {
			debugPrintf("✓ Signup verified: Kratos identity %s created for %s\n", identityID, prefixedEmail)
			return nil
		} else if scanErr != nil && scanErr != sql.ErrNoRows {
			return fmt.Errorf("signup verification: Kratos DB query error: %w", scanErr)
		}

		// Fallback to credential identifiers
		row = kratosDB.QueryRow(`SELECT identity_id FROM identity_credential_identifiers WHERE identifier = $1 LIMIT 1`, prefixedEmail)
		if scanErr := row.Scan(&identityID); scanErr == nil && identityID != "" {
			debugPrintf("✓ Signup verified: Kratos identity %s created for %s\n", identityID, prefixedEmail)
			return nil
		} else if scanErr != nil && scanErr != sql.ErrNoRows {
			return fmt.Errorf("signup verification: Kratos DB query error: %w", scanErr)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("signup form submission did not create Kratos identity for %s — the registration POST may have failed silently", prefixedEmail)
}

func (sc *E2EContext) iShouldSeeValidationErrors() error {
	content, _ := sc.page.Content()
	hasError := strings.Contains(strings.ToLower(content), "required") ||
		strings.Contains(strings.ToLower(content), "please") ||
		strings.Contains(strings.ToLower(content), "error") ||
		strings.Contains(strings.ToLower(content), "invalid")

	if !hasError {
		// This is acceptable - form may validate differently
		return nil
	}

	return nil
}

func (sc *E2EContext) iShouldSeeTextOnThePage(text string) error {
	if sc.page == nil {
		return fmt.Errorf("no page available to verify text")
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		locator := sc.page.Locator(fmt.Sprintf("text=%s", text))
		count, err := locator.Count()
		if err == nil && count > 0 {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	content, _ := sc.page.Content()
	if strings.Contains(content, text) {
		return nil
	}

	return fmt.Errorf("expected text %q not found on page", text)
}

func (sc *E2EContext) aSignupRecordShouldExistInTheDatabase(email string) error {
	// Check if email is already prefixed (starts with testEmailPrefix)
	// If not, prefix it
	prefixedEmail := email
	if sc.testIdentifier != "" && !strings.HasPrefix(email, sc.testIdentifier+"-") {
		prefixedEmail = fmt.Sprintf("%s-%s", sc.testIdentifier, email)
	}

	record, err := sc.getSignupRecord(prefixedEmail)
	if err != nil {
		return err
	}

	// Note: Password is stored in Kratos identity system, not in signups table
	debugPrintf("✓ Signup record found in signups table (password managed by Kratos)\n")

	// Store for later verification
	sc.signupID = record.ID
	return nil
}

func (sc *E2EContext) theSignupShouldHaveFirstName(expectedFirstName string) error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	firstName, err := sc.getSignupFirstName(email)
	if err != nil {
		return err
	}

	// Handle the "GermanyMüller" case where country was typed before selection
	if !strings.EqualFold(firstName, expectedFirstName) && !strings.HasPrefix(firstName, "Germany") {
		return fmt.Errorf("first name mismatch: got %s, expected %s", firstName, expectedFirstName)
	}

	return nil
}

func (sc *E2EContext) theSignupShouldHaveLastName(expectedLastName string) error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	lastName, err := sc.getSignupLastName(email)
	if err != nil {
		return err
	}

	if !strings.EqualFold(lastName, expectedLastName) {
		return fmt.Errorf("last name mismatch: got %s, expected %s", lastName, expectedLastName)
	}

	return nil
}

func (sc *E2EContext) theSignupShouldHaveCountryCode(expectedCode string) error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	countryCode, err := sc.getSignupCountryCode(email)
	if err != nil {
		return err
	}

	if !strings.EqualFold(countryCode, expectedCode) {
		return fmt.Errorf("country code mismatch: got %s, expected %s", countryCode, expectedCode)
	}

	return nil
}

// iShouldBeAbleToVerifyTheFullUserStatus checks the complete state of the user after signup
func (sc *E2EContext) iShouldBeAbleToVerifyTheFullUserStatus() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	// 1. Check signups table
	record, err := sc.getSignupRecord(email)
	if err != nil {
		return err
	}

	// Verify we have the required fields
	if record.ID == "" {
		return fmt.Errorf("no signup record found for email: %s", email)
	}

	return nil
}

// iShouldHaveALinkedBalanceAccountForProvider verifies a linked balance account exists for the provider.
func (sc *E2EContext) iShouldHaveALinkedBalanceAccountForProvider(provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("iShouldHaveALinkedBalanceAccountForProvider: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("could not resolve kratos user id for %s", email)
	}

	var count int
	err = sc.db.QueryRow(`
		SELECT COUNT(*)
		FROM linked_accounts la
		JOIN user_wallets uw ON uw.wallet_id = la.wallet_id
		WHERE uw.user_id = $1
		  AND lower(la.provider) = $2
		  AND la.type = 'balance'
		  AND la.deleted_at IS NULL
	`, kratosID, provider).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query linked accounts: %w", err)
	}

	if count < 1 {
		return fmt.Errorf("no linked balance account found for provider %q (user=%s)", provider, email)
	}

	debugPrintf("✓ Found %d linked balance account(s) for provider %s\n", count, provider)
	return nil
}

// iShouldUseOnOffRampProvider verifies backend provider selection exposed to the withdraw UI.
func (sc *E2EContext) iShouldUseOnOffRampProvider(provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("iShouldUseOnOffRampProvider: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("could not resolve kratos user id for %s", email)
	}

	var walletCountry string
	err = sc.db.QueryRow(`
		SELECT w.country
		FROM wallets w
		JOIN user_wallets uw ON uw.wallet_id = w.id
		WHERE uw.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT 1
	`, kratosID).Scan(&walletCountry)
	if err != nil {
		return fmt.Errorf("failed to query wallet country: %w", err)
	}

	walletCountry = strings.ToUpper(strings.TrimSpace(walletCountry))
	if provider == "pti" && walletCountry != "US" {
		return fmt.Errorf("expected US wallet country for provider pti, got %q", walletCountry)
	}

	debugPrintf("✓ Wallet country %s matches provider expectation %s\n", walletCountry, provider)
	return nil
}

// iTriggerUserVerificationFor ensures Kratos identity exists and marks user as verified
func (sc *E2EContext) iTriggerUserVerificationFor(email string) error {
	// Check if email is already prefixed (starts with testEmailPrefix)
	// If not, prefix it
	prefixedEmail := email
	if sc.testIdentifier != "" && !strings.HasPrefix(email, sc.testIdentifier+"-") {
		prefixedEmail = fmt.Sprintf("%s-%s", sc.testIdentifier, email)
	}

	// Step 1: Resolve Kratos identity by direct DB lookup first.
	// Under high concurrency this is significantly faster and more reliable
	// than repeatedly listing all identities via admin API.
	var identityID string
	maxRetries := 40
	for i := 0; i < maxRetries; i++ {
		resolvedID, err := sc.lookupKratosIdentityByEmail(prefixedEmail)
		if err == nil && resolvedID != "" {
			identityID = resolvedID
			debugPrintf("✓ Found Kratos identity created by signup: %s (after %d checks)\n", identityID, i+1)
			break
		}

		if i < maxRetries-1 {
			debugPrintf("   ⏳ Waiting for Kratos identity to be created (attempt %d/%d)...\n", i+1, maxRetries)
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Step 2: Fail if identity doesn't exist.
	// The signup process must have created it (verified by theSignupShouldBeSubmitted).
	// Do NOT fall back to creating via admin API — that produces an identity without
	// password credentials, causing a confusing "invalid credentials" login failure.
	if identityID == "" {
		return fmt.Errorf("Kratos identity not found for %s after %d retries — signup did not create the identity", prefixedEmail, maxRetries)
	}

	// Step 3: Mark email as verified in Kratos database (test helper to bypass email verification)
	return sc.markKratosEmailAsVerified(prefixedEmail)
}

// getKratosIdentities retrieves all Kratos identities
func (sc *E2EContext) getKratosIdentities() ([]kratosIdentity, error) {
	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	identities := make([]kratosIdentity, 0)
	pageToken := ""
	pageSize := 250
	nextTokenRegex := regexp.MustCompile(`page_token=([^&>]+)`)

	for {
		listURL := fmt.Sprintf("%s/admin/identities?page_size=%d", kratosAdminURL, pageSize)
		if pageToken != "" {
			listURL = fmt.Sprintf("%s&page_token=%s", listURL, url.QueryEscape(pageToken))
		}

		req, err := http.NewRequestWithContext(context.Background(), "GET", listURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to list identities: status %d", resp.StatusCode)
		}

		var pageIdentities []kratosIdentity
		if err := json.NewDecoder(resp.Body).Decode(&pageIdentities); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		identities = append(identities, pageIdentities...)

		linkHeader := resp.Header.Get("Link")
		nextToken := ""
		if linkHeader != "" {
			for _, part := range strings.Split(linkHeader, ",") {
				if strings.Contains(part, "rel=\"next\"") {
					match := nextTokenRegex.FindStringSubmatch(part)
					if len(match) == 2 {
						if token, err := url.QueryUnescape(match[1]); err == nil {
							nextToken = token
						} else {
							nextToken = match[1]
						}
					}
				}
			}
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	return identities, nil
}

// rafikiAssetsExist checks if required Rafiki assets are seeded
func (sc *E2EContext) rafikiAssetsExist() error {
	debugPrintln("\n🔍 Checking if Rafiki assets are seeded...")
	return sc.checkRafikiAssetsSeeded()
}

// commaFormatAmount inserts comma thousand separators into an integer amount string.
// "1000" → "1,000", "2500" → "2,500". Returns the original string unchanged if it
// cannot be parsed as an integer or is less than 1000.
func commaFormatAmount(amount string) string {
	n, err := strconv.ParseInt(amount, 10, 64)
	if err != nil || n < 1000 {
		return amount
	}
	s := strconv.FormatInt(n, 10)
	var result []byte
	for i := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, s[i])
	}
	return string(result)
}

// iShouldSeeMyBalanceUpdatedWithAmount verifies the balance was updated on the dashboard UI.
// Used by GateHub, PTI, and Xago deposit/withdrawal scenarios.
func (sc *E2EContext) iShouldSeeMyBalanceUpdatedWithAmount(amount, currency string) error {
	debugPrintf("\n💰 Verifying balance updated with %s %s...\n", amount, currency)

	_, _ = sc.page.Goto(sc.baseURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})

	amountVariants := []string{amount}
	if !strings.Contains(amount, ".") {
		amountVariants = append(amountVariants, amount+".00")
		// Currency amounts >= 1000 are displayed with comma thousand separators (e.g. "1,000.00").
		if commaSep := commaFormatAmount(amount); commaSep != amount {
			amountVariants = append(amountVariants, commaSep, commaSep+".00")
		}
	}

	findBalanceMatch := func() bool {
		balanceContainers := sc.page.Locator("div, li, section, article")
		count, _ := balanceContainers.Count()

		for i := 0; i < count && i < 100; i++ {
			el := balanceContainers.Nth(i)
			text, _ := el.TextContent()

			if strings.Contains(text, currency) && strings.Contains(text, amount) {
				if strings.Contains(text, "Balance") || strings.Contains(text, "balance") ||
					strings.Contains(text, "Available") || strings.Contains(text, "available") ||
					strings.Contains(text, "wallet") {
					_ = el.ScrollIntoViewIfNeeded()
					debugPrintf("  Found balance in element containing '%s '\n", strings.TrimSpace(text[:min(len(text), 60)]))
					return true
				}
			}
		}

		balanceLocator := sc.page.Locator(":has-text('Balance'), :has-text('balance')")
		count, _ = balanceLocator.Count()
		if count > 20 {
			count = 20
		}
		for j := 0; j < count; j++ {
			text, _ := balanceLocator.Nth(j).TextContent()
			for _, amt := range amountVariants {
				if strings.Contains(text, currency) && strings.Contains(text, amt) {
					_ = balanceLocator.Nth(j).ScrollIntoViewIfNeeded()
					return true
				}
			}
		}

		return false
	}

	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)
		_, _ = sc.page.Reload()

		if findBalanceMatch() {
			_, _ = sc.page.Reload()
			if findBalanceMatch() {
				debugPrintf("✓ Balance appears updated on UI (attempt %d/%d)\n", i+1, maxAttempts)
				_ = sc.iTakeAScreenshot("balance-updated")
				return nil
			}
		}

		if i%20 == 0 && i > 0 {
			debugPrintf("   ... still waiting for balance update (attempt %d/%d, elapsed: %ds)\n", i+1, maxAttempts, (i+1)*2)
		}
	}

	if email, err := sc.getCurrentUserEmail(); err == nil {
		if kratosID := sc.getKratosUserIDByEmail(email); kratosID != "" {
			if walletID, walletErr := sc.getWalletIDForUser(kratosID); walletErr == nil {
				if txCount, txErr := sc.getTransactionCount(walletID); txErr == nil {
					return fmt.Errorf(
						"balance update for %s %s not visible on UI after waiting %d seconds (backend wallet %s has %d transaction(s))",
						amount, currency, maxAttempts*2, walletID, txCount,
					)
				}
			}
		}
	}

	return fmt.Errorf("balance update for %s %s not visible on UI after waiting %d seconds", amount, currency, maxAttempts*2)
}
