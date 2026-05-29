package storage

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
)

// SeedTestUsers creates pre-seeded test users with balances
func SeedTestUsers(store Storage) error {
	return SeedTestUsersWithOrgID(store, "default-org")
}

// SeedTestUsersWithOrgID creates pre-seeded test users with balances and a default organization
func SeedTestUsersWithOrgID(store Storage, defaultOrgID string) error {
	// Test User 1: USD balance
	user1 := &models.User{
		ID:        consts.TestUser1ID,
		Email:     consts.TestUser1Email,
		Activated: true,
		Managed:   true,
		Role:      "user",
		Features:  []string{"wallet", "kyc"},
		KYCState:  consts.KYCStateActionRequired,
		RiskLevel: consts.RiskLevelLow,
	}

	if err := store.CreateUser(user1); err != nil {
		// User might already exist, ignore error
	}

	// Add 10,000 USD balance only if not already seeded
	if bal, _ := store.GetBalance(user1.ID, "USD"); bal == 0 {
		if err := store.AddBalance(user1.ID, "USD", 10000.00); err != nil {
			return err
		}
	}

	// Test User 2: EUR balance
	user2 := &models.User{
		ID:        consts.TestUser2ID,
		Email:     consts.TestUser2Email,
		Activated: true,
		Managed:   true,
		Role:      "user",
		Features:  []string{"wallet", "kyc"},
		KYCState:  consts.KYCStateActionRequired,
		RiskLevel: consts.RiskLevelLow,
	}

	if err := store.CreateUser(user2); err != nil {
		// User might already exist, ignore error
	}

	// Add 10,000 EUR balance only if not already seeded
	if bal, _ := store.GetBalance(user2.ID, "EUR"); bal == 0 {
		if err := store.AddBalance(user2.ID, "EUR", 10000.00); err != nil {
			return err
		}
	}

	xagoGatehubGhOmnibusUser := &models.User{
		ID:        consts.TestXagoGatehubGhOmnibusUserID,
		Email:     consts.TestXagoGatehubGhOmnibusUserEmail,
		Activated: true,
		Managed:   true,
		Role:      "user",
		Features:  []string{"wallet"},
		KYCState:  consts.KYCStateAccepted,
		RiskLevel: consts.RiskLevelLow,
	}

	if err := store.CreateUser(xagoGatehubGhOmnibusUser); err != nil {
		// User might already exist, ignore error
	}

	if bal, _ := store.GetBalance(xagoGatehubGhOmnibusUser.ID, "EUR"); bal == 0 {
		if err := store.AddBalance(xagoGatehubGhOmnibusUser.ID, "EUR", 10000.00); err != nil {
			return err
		}
	}

	// Create default organization
	now := time.Now()
	defaultOrg := &models.Organization{
		ID:        defaultOrgID,
		TwoFAType: "sms",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateOrganization(defaultOrg); err != nil {
		// Only ignore "already exists" errors; propagate real failures (e.g., Redis connectivity)
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create default organization: %w", err)
		}
	}

	return nil
}
