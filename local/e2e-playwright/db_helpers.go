package main

import (
	"database/sql"
	"fmt"
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
