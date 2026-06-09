package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// phoneAllocMu serializes phone number allocation across concurrent goroutines.
var phoneAllocMu sync.Mutex

// phoneCounters tracks the next suffix to allocate per country.
// Initialized from the DB max on first call, then incremented in-memory
// so concurrent goroutines never receive the same number — even before
// Kratos has registered the previous allocation.
var phoneCounters = map[string]int{}

var (
	_ = (*E2EContext).getGatehubWalletIDByEmail
	_ = (*E2EContext).getWalletIDByEmail
	_ = (*E2EContext).getKYCStatusByWalletID
)

const backendDBConnStr = "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"

// ensureDB opens the backend database connection if it is not already open.
func (sc *E2EContext) ensureDB() error {
	if sc.db != nil {
		return nil
	}
	db, err := sql.Open("postgres", backendDBConnStr)
	if err != nil {
		return fmt.Errorf("failed to open backend db: %w", err)
	}
	sc.db = db
	return nil
}

// waitForStableWalletCount polls the backend DB for the number of wallets
// associated with the current test user. It returns when the observed
// count is >= expectedMin for `stableFor` consecutive checks or when
// the timeout is reached.
func (sc *E2EContext) waitForStableWalletCount(expectedMin int, stableFor int, timeout time.Duration) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("waitForStableWalletCount: %w", err)
	}

	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("waitForStableWalletCount: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return fmt.Errorf("waitForStableWalletCount: could not resolve kratos user id for %s", email)
	}

	deadline := time.Now().Add(timeout)
	consecutive := 0

	for time.Now().Before(deadline) {
		var count int
		err := sc.db.QueryRow(`SELECT COUNT(*) FROM user_wallets WHERE user_id = $1`, kratosID).Scan(&count)
		if err != nil {
			// transient DB error — retry
			debugPrintf("   ⚠️  waitForStableWalletCount: query error: %v\n", err)
			consecutive = 0
			time.Sleep(250 * time.Millisecond)
			continue
		}

		debugPrintf("   📊 wallet count for user %s: %d (expect >= %d)\n", kratosID, count, expectedMin)

		if count >= expectedMin {
			consecutive++
			if consecutive >= stableFor {
				return nil
			}
		} else {
			consecutive = 0
		}

		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("waitForStableWalletCount: timeout waiting for >=%d wallets", expectedMin)
}

// getGatehubWalletIDByEmail fetches the GateHub wallet ID (provider_id) for a user by email
// It queries the linked_accounts table to find the GateHub provider_id for the user
func (sc *E2EContext) getGatehubWalletIDByEmail(email string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getGatehubWalletIDByEmail: %w", err)
	}

	// Get the Kratos user ID from the email
	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return "", fmt.Errorf("getGatehubWalletIDByEmail: could not resolve kratos user id for %s", email)
	}

	debugPrintf("   📋 Looking up GateHub wallet ID for email: %s (kratos ID: %s)\n", email, kratosID)

	// Query the linked_accounts table to get the GateHub provider_id
	var gateHubWalletID string
	err := sc.db.QueryRow(`
		SELECT provider_id FROM linked_accounts 
		WHERE wallet_id IN (
			SELECT wallet_id FROM user_wallets WHERE user_id = $1
		)
		AND provider = 'gatehub'
		LIMIT 1
	`, kratosID).Scan(&gateHubWalletID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getGatehubWalletIDByEmail: no GateHub linked account found for user %s", email)
		}
		return "", fmt.Errorf("getGatehubWalletIDByEmail: database error: %w", err)
	}

	debugPrintf("   ✓ Found GateHub wallet ID: %s\n", gateHubWalletID)
	return gateHubWalletID, nil
}

// getGatehubUserIDByEmail fetches the GateHub managed user ID (external_id) for a user by email.
// It queries the gatehub_users table using the user's wallet associations.
func (sc *E2EContext) getGatehubUserIDByEmail(email string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getGatehubUserIDByEmail: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return "", fmt.Errorf("getGatehubUserIDByEmail: could not resolve kratos user id for %s", email)
	}

	debugPrintf("   📋 Looking up GateHub user ID for email: %s (kratos ID: %s)\n", email, kratosID)

	var gatehubUserID string
	err := sc.db.QueryRow(`
		SELECT gu.external_id
		FROM gatehub_users gu
		JOIN user_wallets uw ON uw.wallet_id = gu.wallet_id
		WHERE uw.user_id = $1
		LIMIT 1
	`, kratosID).Scan(&gatehubUserID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getGatehubUserIDByEmail: no GateHub user found for %s", email)
		}
		return "", fmt.Errorf("getGatehubUserIDByEmail: database error: %w", err)
	}

	debugPrintf("   ✓ Found GateHub user ID: %s\n", gatehubUserID)
	return gatehubUserID, nil
}

// SignupRecord holds signup data from the database
type SignupRecord struct {
	ID          string
	Email       string
	FirstName   string
	LastName    string
	CountryCode string
	UserID      sql.NullString
}

// getSignupRecord retrieves signup record for an email address
func (sc *E2EContext) getSignupRecord(email string) (*SignupRecord, error) {
	if err := sc.ensureDB(); err != nil {
		return nil, fmt.Errorf("getSignupRecord: %w", err)
	}

	var record SignupRecord
	err := sc.db.QueryRow("SELECT id, email, first_name, last_name, country_code, user_id FROM signups WHERE email = $1", email).
		Scan(&record.ID, &record.Email, &record.FirstName, &record.LastName, &record.CountryCode, &record.UserID)

	if err == sql.ErrNoRows {
		// Try waitlist_signups
		err = sc.db.QueryRow("SELECT id, email, first_name, last_name, country_code, user_id FROM waitlist_signups WHERE email = $1", email).
			Scan(&record.ID, &record.Email, &record.FirstName, &record.LastName, &record.CountryCode, &record.UserID)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("getSignupRecord: no signup record found for %s", email)
			}
			return nil, fmt.Errorf("getSignupRecord: failed to query waitlist_signups: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("getSignupRecord: failed to query signups: %w", err)
	}

	return &record, nil
}

// getSignupFirstName retrieves the first name from a signup record
func (sc *E2EContext) getSignupFirstName(email string) (string, error) {
	record, err := sc.getSignupRecord(email)
	if err != nil {
		return "", err
	}
	return record.FirstName, nil
}

// getSignupLastName retrieves the last name from a signup record
func (sc *E2EContext) getSignupLastName(email string) (string, error) {
	record, err := sc.getSignupRecord(email)
	if err != nil {
		return "", err
	}
	return record.LastName, nil
}

// getSignupCountryCode retrieves the country code from a signup record
func (sc *E2EContext) getSignupCountryCode(email string) (string, error) {
	record, err := sc.getSignupRecord(email)
	if err != nil {
		return "", err
	}
	return record.CountryCode, nil
}

// markKratosEmailAsVerified marks an email as verified directly in the Kratos DB.
func (sc *E2EContext) markKratosEmailAsVerified(email string) error {
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		return fmt.Errorf("markKratosEmailAsVerified: could not connect to Kratos DB: %w", err)
	}
	defer kratosDB.Close()

	result, err := kratosDB.Exec(`
		UPDATE identity_verifiable_addresses
		SET verified = TRUE, verified_at = NOW(), status = 'completed', updated_at = NOW()
		WHERE value = $1
	`, email)
	if err != nil {
		return fmt.Errorf("markKratosEmailAsVerified: query failed for %s: %w", email, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("markKratosEmailAsVerified: could not check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("markKratosEmailAsVerified: no verifiable address found for %s", email)
	}

	debugPrintf("✓ Marked email as verified in Kratos: %s\n", email)
	return nil
}

// markKratosPhoneAsVerified marks the current user's phone as verified in Kratos traits.
func (sc *E2EContext) markKratosPhoneAsVerified(email string) error {
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: could not connect to Kratos DB: %w", err)
	}
	defer kratosDB.Close()

	identityID, err := sc.lookupKratosIdentityByEmail(email)
	if err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: failed to resolve identity for %s: %w", email, err)
	}

	var traitsRaw []byte
	if err := kratosDB.QueryRow(`SELECT traits FROM identities WHERE id = $1`, identityID).Scan(&traitsRaw); err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: failed to load traits for %s: %w", identityID, err)
	}

	var traits map[string]interface{}
	if err := json.Unmarshal(traitsRaw, &traits); err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: failed to decode traits for %s: %w", identityID, err)
	}

	phone, hasPhone := traits["phone"].(string)
	if !hasPhone || strings.TrimSpace(phone) == "" {
		return fmt.Errorf("markKratosPhoneAsVerified: no phone trait found for %s", identityID)
	}

	traits["phoneVerified"] = true
	updatedTraits, err := json.Marshal(traits)
	if err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: failed to encode traits for %s: %w", identityID, err)
	}

	result, err := kratosDB.Exec(`
		UPDATE identities
		SET traits = $2, updated_at = NOW()
		WHERE id = $1
	`, identityID, updatedTraits)
	if err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: failed to update traits for %s: %w", identityID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("markKratosPhoneAsVerified: could not check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("markKratosPhoneAsVerified: no identity row updated for %s", identityID)
	}

	debugPrintf("✓ Marked phone as verified in Kratos: %s (%s)\n", phone, identityID)
	return nil
}

// lookupKratosIdentityByEmail looks up Kratos identity in the database
func (sc *E2EContext) lookupKratosIdentityByEmail(email string) (string, error) {
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		return "", fmt.Errorf("lookupKratosIdentityByEmail: could not connect to Kratos DB: %w", err)
	}
	defer kratosDB.Close()

	var identityID string

	// Prefer verifiable addresses when available.
	row := kratosDB.QueryRow(`SELECT identity_id FROM identity_verifiable_addresses WHERE value = $1 LIMIT 1`, email)
	if err := row.Scan(&identityID); err == nil {
		return identityID, nil
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("lookupKratosIdentityByEmail: verifiable_addresses query error: %w", err)
	}

	// Fallback to credential identifiers. This catches identities that exist
	// but have not populated verifiable addresses yet.
	row = kratosDB.QueryRow(`SELECT identity_id FROM identity_credential_identifiers WHERE identifier = $1 LIMIT 1`, email)
	if err := row.Scan(&identityID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("lookupKratosIdentityByEmail: identity not found for %s", email)
		}
		return "", fmt.Errorf("lookupKratosIdentityByEmail: credential_identifiers query error: %w", err)
	}

	return identityID, nil
}

// checkRafikiAssetsSeeded verifies that required Rafiki assets are seeded in the database
func (sc *E2EContext) checkRafikiAssetsSeeded() error {
	requiredAssets := []string{"EUR", "USD"}

	// Connect to Rafiki database
	rafikiConnStr := os.Getenv("RAFIKI_DB_URL")
	if rafikiConnStr == "" {
		rafikiConnStr = "host=localhost port=5432 user=postgres password=postgres dbname=rafiki_backend sslmode=disable"
	}

	rafikiDB, err := sql.Open("postgres", rafikiConnStr)
	if err != nil {
		return fmt.Errorf("❌ checkRafikiAssetsSeeded: could not connect to Rafiki database: %w\n\n"+
			"Ensure Rafiki is running:\n"+
			"  cd local\n"+
			"  docker compose up -d rafiki\n", err)
	}
	defer rafikiDB.Close()

	// Test connection
	if err := rafikiDB.Ping(); err != nil {
		return fmt.Errorf("❌ checkRafikiAssetsSeeded: could not ping Rafiki database: %w\n\n"+
			"Ensure Rafiki database is accessible:\n"+
			"  docker compose ps postgres\n", err)
	}

	// Query for assets
	query := `SELECT code, scale FROM assets WHERE code IN ('EUR', 'USD') ORDER BY code`
	rows, err := rafikiDB.Query(query)
	if err != nil {
		return fmt.Errorf("❌ checkRafikiAssetsSeeded: could not query Rafiki assets: %w\n\n"+
			"This may indicate Rafiki is not properly initialized.\n"+
			"Run the seed script:\n"+
			"  cd local\n"+
			"  ./scripts/local-dev-tool rafiki\n", err)
	}
	defer rows.Close()

	foundAssets := make(map[string]int)
	for rows.Next() {
		var code string
		var scale int
		if err := rows.Scan(&code, &scale); err != nil {
			continue
		}
		foundAssets[code] = scale
		debugPrintf("   ✓ Found asset: %s (scale: %d)\n", code, scale)
	}

	// Check which assets are missing
	missingAssets := []string{}
	for _, asset := range requiredAssets {
		if _, found := foundAssets[asset]; !found {
			missingAssets = append(missingAssets, asset)
		}
	}

	if len(missingAssets) > 0 {
		return fmt.Errorf("\n❌ RAFIKI ASSETS MISSING: %v\n\n"+
			"The following assets must be seeded in Rafiki before running tests:\n"+
			"  Missing: %s\n\n"+
			"To fix this:\n"+
			"  1. Run the Rafiki seed script:\n"+
			"     cd local\n"+
			"     ./scripts/local-dev-tool rafiki\n\n"+
			"  2. Verify assets were created:\n"+
			"     docker compose exec -T postgres psql -U postgres -d rafiki_backend -c \\\n"+
			"       \"SELECT code, scale FROM assets ORDER BY code;\"\n\n"+
			"  3. Expected output should include:\n"+
			"     code | scale\n"+
			"     -----+-------\n"+
			"     EUR  |     2\n"+
			"     USD  |     2\n",
			missingAssets, strings.Join(missingAssets, ", "))
	}

	debugPrintf("   ✓ All required Rafiki assets present: %v\n", requiredAssets)
	return nil
}

// getUserIDFromSignup retrieves the user_id (Kratos ID) from signups table by email
func (sc *E2EContext) getUserIDFromSignup(email string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getUserIDFromSignup: %w", err)
	}

	var kratosUserID string
	err := sc.db.QueryRow("SELECT user_id FROM signups WHERE email = $1", email).Scan(&kratosUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getUserIDFromSignup: user not found in signups for %s", email)
		}
		return "", fmt.Errorf("getUserIDFromSignup: query error: %w", err)
	}

	return kratosUserID, nil
}

// getWalletIDForUser retrieves the wallet_id for a given user_id
func (sc *E2EContext) getWalletIDForUser(kratosUserID string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getWalletIDForUser: %w", err)
	}

	var walletID string
	err := sc.db.QueryRow("SELECT wallet_id FROM user_wallets WHERE user_id = $1", kratosUserID).Scan(&walletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getWalletIDForUser: wallet not found for user %s", kratosUserID)
		}
		return "", fmt.Errorf("getWalletIDForUser: query error: %w", err)
	}

	return walletID, nil
}

// getTransactionCount retrieves the transaction count for a wallet
func (sc *E2EContext) getTransactionCount(walletID string) (int, error) {
	if err := sc.ensureDB(); err != nil {
		return 0, fmt.Errorf("getTransactionCount: %w", err)
	}

	var count int
	err := sc.db.QueryRow("SELECT COUNT(*) FROM transactions WHERE wallet_id = $1", walletID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getTransactionCount: query error: %w", err)
	}

	return count, nil
}

// getWalletDetailsForUser retrieves wallet details for a user
type WalletDetails struct {
	ID            string
	Name          string
	WalletAddress string
}

func (sc *E2EContext) getWalletDetailsForUser(kratosUserID string) (*WalletDetails, error) {
	if err := sc.ensureDB(); err != nil {
		return nil, fmt.Errorf("getWalletDetailsForUser: %w", err)
	}

	var walletID, walletName, walletAddress sql.NullString
	err := sc.db.QueryRow(`
		SELECT w.id, w.name, (SELECT url FROM wallet_addresses WHERE wallet_id = w.id LIMIT 1)
		FROM wallets w
		JOIN user_wallets uw ON w.id = uw.wallet_id
		WHERE uw.user_id = $1
		ORDER BY w.created_at DESC LIMIT 1
	`, kratosUserID).Scan(&walletID, &walletName, &walletAddress)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("getWalletDetailsForUser: no wallet found for user %s", kratosUserID)
		}
		return nil, fmt.Errorf("getWalletDetailsForUser: query error: %w", err)
	}

	return &WalletDetails{
		ID:            walletID.String,
		Name:          walletName.String,
		WalletAddress: walletAddress.String,
	}, nil
}

// getUserWalletCount returns the count of wallets for a user
func (sc *E2EContext) getUserWalletCount(kratosUserID string) (int, error) {
	if err := sc.ensureDB(); err != nil {
		return 0, fmt.Errorf("getUserWalletCount: %w", err)
	}

	var count int
	err := sc.db.QueryRow(`SELECT COUNT(*) FROM user_wallets WHERE user_id = $1`, kratosUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getUserWalletCount: query error: %w", err)
	}

	return count, nil
}

// allocatePhoneNumber returns a phone number for the given country that does not
// already exist in the Kratos identity_credential_identifiers table.
// It is safe for concurrent use: a process-level mutex ensures only one goroutine
// runs the allocation logic at a time, and an in-memory counter ensures each call
// returns a distinct number even before Kratos has registered the previous one.
func allocatePhoneNumber(country string) (string, error) {
	phoneAllocMu.Lock()
	defer phoneAllocMu.Unlock()

	prefix, pattern, numDigits := phoneRangeForCountry(country)
	// Key counters by the resolved allocation range, not the raw country input,
	// so aliases and fallback countries share the same counter state.
	key := pattern

	// On first call for this country, seed the counter from the DB so we
	// don't collide with phones registered by previous test runs.
	if _, ok := phoneCounters[key]; !ok {
		kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
		kratosDB, err := sql.Open("postgres", kratosConnStr)
		if err != nil {
			return "", fmt.Errorf("allocatePhoneNumber: could not connect to Kratos DB: %w", err)
		}
		defer kratosDB.Close()

		var maxSuffix int
		err = kratosDB.QueryRow(`
			SELECT COALESCE(MAX((regexp_match(identifier, $1))[1]::int), -1)
			FROM identity_credential_identifiers
			WHERE identifier ~ $2
		`, pattern, pattern).Scan(&maxSuffix)
		if err != nil {
			return "", fmt.Errorf("allocatePhoneNumber: failed to query max phone for %s: %w", country, err)
		}

		if maxSuffix >= 0 {
			phoneCounters[key] = maxSuffix + 1
		} else {
			phoneCounters[key] = 0
		}
	}

	limit := 1
	for i := 0; i < numDigits; i++ {
		limit *= 10
	}

	suffix := phoneCounters[key]
	if suffix >= limit {
		return "", fmt.Errorf("allocatePhoneNumber: exhausted phone range for country %s", country)
	}
	phoneCounters[key] = suffix + 1

	return fmt.Sprintf("%s%0*d", prefix, numDigits, suffix), nil
}

// phoneRangeForCountry returns (prefix, regexPattern, suffixDigits) for the
// allocation range used in each country.
func phoneRangeForCountry(country string) (prefix string, pattern string, digits int) {
	switch strings.ToLower(country) {
	case "south africa":
		return "+27710", `^\+27710([0-9]{6})$`, 6
	case "united states", "usa", "us":
		return "+1202555", `^\+1202555([0-9]{4})$`, 4
	default:
		// Germany / fallback
		return "+491700", `^\+491700([0-9]{6})$`, 6
	}
}

// getWalletIDByEmail looks up the wallet ID for a user by their email address
func (sc *E2EContext) getWalletIDByEmail(email string) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("getWalletIDByEmail: %w", err)
	}

	kratosID := sc.getKratosUserIDByEmail(email)
	if kratosID == "" {
		return "", fmt.Errorf("getWalletIDByEmail: could not resolve kratos user id for %s", email)
	}

	var walletID string
	err := sc.db.QueryRow(`SELECT wallet_id FROM user_wallets WHERE user_id = $1 LIMIT 1`, kratosID).Scan(&walletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getWalletIDByEmail: no wallet found for user %s", email)
		}
		return "", fmt.Errorf("getWalletIDByEmail: query error: %w", err)
	}

	debugPrintf("   📋 Found wallet ID %s for email %s\n", walletID, email)
	return walletID, nil
}

// getKYCStatusByWalletID looks up the KYC status for a wallet
func (sc *E2EContext) getKYCStatusByWalletID(walletID string) (int, error) {
	if err := sc.ensureDB(); err != nil {
		return -1, fmt.Errorf("getKYCStatusByWalletID: %w", err)
	}

	var status int
	err := sc.db.QueryRow(`SELECT status FROM wallet_kyc_status WHERE wallet_id = $1`, walletID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Unknown/not started
		}
		return -1, fmt.Errorf("getKYCStatusByWalletID: query error: %w", err)
	}

	debugPrintf("   📋 KYC status for wallet %s: %d\n", walletID, status)
	return status, nil
}

// AgreementSignature mirrors a row in the agreement_signatures table.
type AgreementSignature struct {
	ID                      string
	AgreementID             string
	UserID                  string
	IPAddress               sql.NullString
	LastNotifiedAgreementID sql.NullString
}

// getAgreementSignaturesForUser returns every agreement_signatures row for the given Kratos user ID.
func (sc *E2EContext) getAgreementSignaturesForUser(userID string) ([]AgreementSignature, error) {
	if err := sc.ensureDB(); err != nil {
		return nil, fmt.Errorf("getAgreementSignaturesForUser: %w", err)
	}

	rows, err := sc.db.Query(
		`SELECT id, agreement_id, user_id, ip_address, last_notified_agreement_id
		 FROM agreement_signatures WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("getAgreementSignaturesForUser: query error: %w", err)
	}
	defer rows.Close()

	var sigs []AgreementSignature
	for rows.Next() {
		var s AgreementSignature
		if err := rows.Scan(&s.ID, &s.AgreementID, &s.UserID, &s.IPAddress, &s.LastNotifiedAgreementID); err != nil {
			return nil, fmt.Errorf("getAgreementSignaturesForUser: scan error: %w", err)
		}
		sigs = append(sigs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getAgreementSignaturesForUser: rows error: %w", err)
	}
	return sigs, nil
}

// insertAgreement inserts a new agreements row with notified=false so that the
// agreement-change-notify workflow trigger will pick it up. If the row already
// exists (e.g. a prior test run with a colliding id), every column is refreshed
// — otherwise a stale name would silently redirect the workflow at a different
// agreement than the one the test thinks it published.
func (sc *E2EContext) insertAgreement(id, name, version, content string) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("insertAgreement: %w", err)
	}
	_, err := sc.db.Exec(
		`INSERT INTO agreements (id, name, version, content, notified) VALUES ($1, $2, $3, $4, false)
		 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, version = EXCLUDED.version, content = EXCLUDED.content, notified = false`,
		id, name, version, content,
	)
	if err != nil {
		return fmt.Errorf("insertAgreement: %w", err)
	}
	return nil
}

// cleanupTestAgreement removes the agreement row injected by a scenario, along
// with every agreement_signatures reference. Schema (db/schema.hcl:37-48)
// defines TWO FKs from agreement_signatures back to agreements
// (fk_last_notified_agreement and fk_agreement), both with on_delete=NO_ACTION,
// so both must be cleared before the agreements DELETE — otherwise a future
// test that signs the test agreement leaves the row leaked with notified=false.
// Errors are best-effort: scenario teardown swallows them.
func (sc *E2EContext) cleanupTestAgreement(id string) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("cleanupTestAgreement: %w", err)
	}
	if _, err := sc.db.Exec(`UPDATE agreement_signatures SET last_notified_agreement_id = NULL WHERE last_notified_agreement_id = $1`, id); err != nil {
		return fmt.Errorf("cleanupTestAgreement: clear last_notified references: %w", err)
	}
	if _, err := sc.db.Exec(`DELETE FROM agreement_signatures WHERE agreement_id = $1`, id); err != nil {
		return fmt.Errorf("cleanupTestAgreement: clear agreement_id references: %w", err)
	}
	if _, err := sc.db.Exec(`DELETE FROM agreements WHERE id = $1`, id); err != nil {
		return fmt.Errorf("cleanupTestAgreement: delete agreement: %w", err)
	}
	return nil
}

// pollKratosUserID waits for signups.user_id to be populated by gRPC CompleteSignup.
// CompleteSignup is invoked by the frontend AFTER Kratos identity creation, so
// theSignupShouldBeSubmitted (which only waits for the Kratos row) can return
// while signups.user_id is still NULL. Uses sql.NullString so a NULL scan does
// not panic with 'converting NULL to string is unsupported' (see e2e/AGENTS.md).
func (sc *E2EContext) pollKratosUserID(email string, timeout time.Duration) (string, error) {
	if err := sc.ensureDB(); err != nil {
		return "", fmt.Errorf("pollKratosUserID: %w", err)
	}
	deadline := time.Now().Add(timeout)
	for {
		var id sql.NullString
		err := sc.db.QueryRow("SELECT user_id FROM signups WHERE email = $1", email).Scan(&id)
		if err == nil && id.Valid && id.String != "" {
			return id.String, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return "", fmt.Errorf("pollKratosUserID: query error for %s: %w", email, err)
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("pollKratosUserID: signups.user_id remained NULL for %s after %v — grpc.CompleteSignup likely did not run server-side", email, timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// waitForAgreementSignature polls until the user has a signature for the given
// agreement_id, or until timeout. Covers two consecutive race windows: (1) the
// signups.user_id UPDATE in CompleteSignup, and (2) the agreement_signatures
// INSERT that runs immediately after within the same gRPC call (a separate DB
// transaction, so visibility is staggered by milliseconds under load).
func (sc *E2EContext) waitForAgreementSignature(email, agreementID string, timeout time.Duration) error {
	if err := sc.ensureDB(); err != nil {
		return fmt.Errorf("waitForAgreementSignature: %w", err)
	}
	deadline := time.Now().Add(timeout)
	var lastUserID string
	var lastSeen []string
	for {
		// Reset diagnostics at the top of each iteration so the eventual
		// timeout error reflects the current DB state, not a stale snapshot
		// from a prior iteration that succeeded then degraded.
		lastUserID = ""
		lastSeen = lastSeen[:0]
		var id sql.NullString
		if err := sc.db.QueryRow("SELECT user_id FROM signups WHERE email = $1", email).Scan(&id); err == nil && id.Valid && id.String != "" {
			lastUserID = id.String
			sigs, sigErr := sc.getAgreementSignaturesForUser(id.String)
			if sigErr == nil {
				for _, s := range sigs {
					if s.AgreementID == agreementID {
						debugPrintf("   ✓ Found agreement signature %s for user %s\n", agreementID, id.String)
						return nil
					}
					lastSeen = append(lastSeen, s.AgreementID)
				}
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("waitForAgreementSignature: timed out after %v waiting for signature %q for %s (resolved user_id=%q, last seen signatures: %v)",
				timeout, agreementID, email, lastUserID, lastSeen)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
