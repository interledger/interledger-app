package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

// waitForStableWalletCount polls the backend DB for the number of wallets
// associated with the current test user. It returns when the observed
// count is >= expectedMin for `stableFor` consecutive checks or when
// the timeout is reached.
func (sc *E2EContext) waitForStableWalletCount(expectedMin int, stableFor int, timeout time.Duration) error {
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("waitForStableWalletCount: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("getGatehubWalletIDByEmail: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("getGatehubUserIDByEmail: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("getSignupRecord: failed to open db: %w", err)
		}
		sc.db = db
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

// markKratosEmailAsVerified marks an email as verified in the Kratos database
func (sc *E2EContext) markKratosEmailAsVerified(email string) error {
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		debugPrintf("⚠️  markKratosEmailAsVerified: could not connect to Kratos database: %v\n", err)
		return nil
	}
	defer kratosDB.Close()

	_, err = kratosDB.Exec(`
		UPDATE identity_verifiable_addresses 
		SET verified = TRUE, verified_at = NOW()
		WHERE value = $1
	`, email)

	if err != nil {
		debugPrintf("⚠️  markKratosEmailAsVerified: could not mark email as verified: %v\n", err)
		return nil
	}

	debugPrintf("✓ Marked email as verified in Kratos: %s\n", email)
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
	row := kratosDB.QueryRow(`SELECT identity_id FROM identity_verifiable_addresses WHERE value = $1 LIMIT 1`, email)
	if err := row.Scan(&identityID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("lookupKratosIdentityByEmail: identity not found for %s", email)
		}
		return "", fmt.Errorf("lookupKratosIdentityByEmail: query error: %w", err)
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("getUserIDFromSignup: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("getWalletIDForUser: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return 0, fmt.Errorf("getTransactionCount: failed to open db: %w", err)
		}
		sc.db = db
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("getWalletDetailsForUser: failed to open db: %w", err)
		}
		sc.db = db
	}

	var walletID, walletName, walletAddress sql.NullString
	err := sc.db.QueryRow(`
		SELECT w.id, w.name, w.wallet_address 
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
	if sc.db == nil {
		connStr := "host=localhost port=5432 user=postgres password=postgres dbname=backend sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return 0, fmt.Errorf("getUserWalletCount: failed to open db: %w", err)
		}
		sc.db = db
	}

	var count int
	err := sc.db.QueryRow(`SELECT COUNT(*) FROM user_wallets WHERE user_id = $1`, kratosUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getUserWalletCount: query error: %w", err)
	}

	return count, nil
}

// getNextAvailableUSPhoneNumber returns the next available Kratos-safe US test phone number.
// Format: +1202555xxxx (xxxx from 0000 to 9999).
func (sc *E2EContext) getNextAvailableUSPhoneNumber() (string, error) {
	kratosConnStr := "host=localhost port=5432 user=postgres password=postgres dbname=kratos sslmode=disable"
	kratosDB, err := sql.Open("postgres", kratosConnStr)
	if err != nil {
		return "", fmt.Errorf("getNextAvailableUSPhoneNumber: could not connect to Kratos DB: %w", err)
	}
	defer kratosDB.Close()

	var maxSuffix int
	err = kratosDB.QueryRow(`
		SELECT COALESCE(MAX((regexp_match(identifier, '^\\+1202555([0-9]{4})$'))[1]::int), -1)
		FROM identity_credential_identifiers
		WHERE identifier ~ '^\\+1202555[0-9]{4}$'
	`).Scan(&maxSuffix)
	if err != nil {
		return "", fmt.Errorf("getNextAvailableUSPhoneNumber: failed to query max existing phone: %w", err)
	}

	start := 0
	if maxSuffix >= 0 {
		start = maxSuffix + 1
	}

	for i := start; i <= 9999; i++ {
		candidate := fmt.Sprintf("+1202555%04d", i)

		var exists bool
		err = kratosDB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM identity_credential_identifiers WHERE identifier = $1
			)
		`, candidate).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("getNextAvailableUSPhoneNumber: failed existence check for %s: %w", candidate, err)
		}

		if !exists {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("getNextAvailableUSPhoneNumber: exhausted +1202555xxxx range")
}
