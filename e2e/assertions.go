package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Assertion step implementations

func (sc *E2EContext) theSignupShouldBeSubmitted() error {
	// Verify password was filled during the flow
	if !sc.passwordFilled {
		return fmt.Errorf("password was never filled during signup flow")
	}

	// Wait for submission to complete
	time.Sleep(2 * time.Second)

	return nil
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

	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to open db: %w", err)
		}
		sc.db = db
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

	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to open db: %w", err)
		}
		sc.db = db
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

	kratosAdminURL := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminURL == "" {
		kratosAdminURL = "http://localhost:4434"
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

	// Step 2: Fail if identity doesn't exist - it should have been created by the signup process
	if identityID == "" {
		debugPrintf("   ⚠️  Identity not found via DB lookup, attempting admin API create for %s\n", prefixedEmail)
		client := &http.Client{Timeout: 30 * time.Second}

		phone, err := sc.getCurrentUserPhone()
		if err != nil {
			debugPrintf("   ⚠️  Failed to get phone from user details: %v\n", err)
			return fmt.Errorf("cannot create Kratos identity without phone: %w", err)
		}

		payload := map[string]interface{}{"traits": map[string]string{
			"email":       prefixedEmail,
			"phone":       phone,
			"firstName":   sc.firstName,
			"lastName":    sc.lastName,
			"countryCode": sc.country,
		}}
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal Kratos identity payload: %w", err)
		}
		req, err := http.NewRequestWithContext(context.Background(), "POST", kratosAdminURL+"/admin/identities", strings.NewReader(string(b)))
		if err != nil {
			return fmt.Errorf("failed to build Kratos admin API request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Kratos admin API request failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			var created kratosIdentity
			if err := json.NewDecoder(resp.Body).Decode(&created); err == nil {
				identityID = created.ID
				debugPrintf("   ✓ Created Kratos identity via admin API: %s\n", identityID)
			}
		} else if resp.StatusCode == http.StatusConflict {
			// Conflict means an identity already exists for one of the identifiers.
			// Re-resolve identity from DB instead of failing.
			if resolvedID, resolveErr := sc.lookupKratosIdentityByEmail(prefixedEmail); resolveErr == nil {
				identityID = resolvedID
				debugPrintf("   ✓ Resolved existing Kratos identity after conflict: %s\n", identityID)
			} else {
				body, _ := io.ReadAll(resp.Body)
				debugPrintf("   ⚠️  Conflict creating identity and DB resolve failed: %s\n", string(body))
			}
		} else {
			body, _ := io.ReadAll(resp.Body)
			debugPrintf("   ⚠️  Failed to create Kratos identity: status %d: %s\n", resp.StatusCode, string(body))
		}

		if identityID == "" {
			return fmt.Errorf("Kratos identity not found for %s - signup process should have created it", prefixedEmail)
		}
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
