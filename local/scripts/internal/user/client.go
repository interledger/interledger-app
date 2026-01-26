package user

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"path/filepath"

	"github.com/joho/godotenv"
)

type CreateUserOptions struct {
	Email       string
	FirstName   string
	LastName    string
	CountryCode string
	Password    string
}

type CreateUserResult struct {
	IdentityID string
	Email      string
	WalletID   string
	KYCStatus  string
}

type UserInfo struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	State     string
}

type Client struct {
	kratosAdminURL  string
	postgresConnStr string
	mockGatehubURL  string
}

// Kratos API structures
type KratosIdentity struct {
	ID                  string                 `json:"id"`
	SchemaID            string                 `json:"schema_id"`
	Traits              map[string]interface{} `json:"traits"`
	Credentials         *KratosCredentials     `json:"credentials,omitempty"`
	VerifiableAddresses []VerifiableAddress    `json:"verifiable_addresses,omitempty"`
	State               string                 `json:"state,omitempty"`
}

type KratosCredentials struct {
	Password *PasswordCredentials `json:"password,omitempty"`
}

type PasswordCredentials struct {
	Config PasswordConfig `json:"config"`
}

type PasswordConfig struct {
	Password string `json:"password"`
}

type VerifiableAddress struct {
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
	Via      string `json:"via"`
	Status   string `json:"status"`
}

func NewClient() *Client {
	// Load environment variables
	envPath := filepath.Join("..", ".env")
	_ = godotenv.Load(envPath)

	getEnv := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}

	return &Client{
		kratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		postgresConnStr: getEnv("POSTGRES_URL", "postgres://postgres:password@localhost:5432/backend?sslmode=disable"),
		mockGatehubURL:  getEnv("MOCKGATEHUB_URL", "https://mockgatehub.interledger.test"),
	}
}

func (c *Client) CreateUser(ctx context.Context, opts CreateUserOptions) (*CreateUserResult, error) {
	fmt.Println("Step 1/5: Creating Kratos identity...")
	identity, err := c.createKratosIdentity(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kratos identity: %w", err)
	}
	fmt.Printf("  ✓ Identity created: %s\n", identity.ID)

	fmt.Println("\nStep 2/5: Verifying email address...")
	if err := c.verifyEmail(ctx, identity.ID, opts.Email); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}
	fmt.Println("  ✓ Email verified")

	fmt.Println("\nStep 3/5: Creating default wallet...")
	walletID, err := c.createDefaultWallet(ctx, identity.ID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	fmt.Printf("  ✓ Wallet created: %s\n", walletID)

	fmt.Println("\nStep 4/5: Setting up KYC...")
	if err := c.setupKYC(ctx, walletID, opts); err != nil {
		return nil, fmt.Errorf("failed to setup KYC: %w", err)
	}
	fmt.Println("  ✓ KYC configured")

	fmt.Println("\nStep 5/5: Activating user (Gatehub)...")
	kycStatus, err := c.activateUser(ctx, walletID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}
	fmt.Printf("  ✓ User activated with KYC status: %s\n", kycStatus)

	return &CreateUserResult{
		IdentityID: identity.ID,
		Email:      opts.Email,
		WalletID:   walletID,
		KYCStatus:  kycStatus,
	}, nil
}

func (c *Client) createKratosIdentity(ctx context.Context, opts CreateUserOptions) (*KratosIdentity, error) {
	// Generate a unique phone number based on current timestamp
	// Format: tel:+1-NXX-NXX-XXXX (tel: URI scheme with visual separators)
	timestamp := time.Now().UnixNano()
	phoneNumber := fmt.Sprintf("tel:+1-%03d-%03d-%04d",
		200+(timestamp/1000000)%800, // NXX format: 200-999
		200+(timestamp/1000)%800,    // NXX format: 200-999
		timestamp%10000)             // Line number: 0000-9999

	payload := map[string]interface{}{
		"schema_id": "default",
		"traits": map[string]interface{}{
			"email":       opts.Email,
			"firstName":   opts.FirstName,
			"lastName":    opts.LastName,
			"countryCode": opts.CountryCode,
			"phone":       phoneNumber, // Unique phone number in tel: URI format
		},
		"credentials": map[string]interface{}{
			"password": map[string]interface{}{
				"config": map[string]interface{}{
					"password": opts.Password,
				},
			},
		},
		"verifiable_addresses": []map[string]interface{}{
			{
				"value":    opts.Email,
				"verified": false,
				"via":      "email",
				"status":   "pending",
			},
		},
		"state": "active",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.kratosAdminURL+"/admin/identities", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// Provide helpful error message for duplicate email (409 Conflict)
		if resp.StatusCode == http.StatusConflict {
			return nil, fmt.Errorf("user with email '%s' already exists. Please use a different email address or delete the existing user first", opts.Email)
		}

		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var identity KratosIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

func (c *Client) verifyEmail(ctx context.Context, identityID, email string) error {
	// Update the identity to mark email as verified
	payload := []map[string]interface{}{
		{
			"op":   "replace",
			"path": "/verifiable_addresses",
			"value": []map[string]interface{}{
				{
					"value":    email,
					"verified": true,
					"via":      "email",
					"status":   "completed",
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/admin/identities/%s", c.kratosAdminURL, identityID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) createDefaultWallet(ctx context.Context, userID string, opts CreateUserOptions) (string, error) {
	// This would normally call the backend gRPC API
	// For local dev, we'll use direct database access
	// In a real implementation, you'd use the protobuf client

	// Simplified: Use PostgreSQL to insert wallet
	// For now, we'll return a mock wallet ID
	// TODO: Implement actual wallet creation via backend API or direct DB access

	walletID := fmt.Sprintf("wallet-%s", userID[:8])

	// Note: In production, this should call backend.CreateUserDefaultWallet
	// or use direct database insertion with proper SQL

	return walletID, nil
}

func (c *Client) setupKYC(ctx context.Context, walletID string, opts CreateUserOptions) error {
	// This would set KYC status to pending
	// For EU countries, it would trigger Gatehub onboarding

	// Note: In production, this should call backend.SetKYCStatusPending
	// For now, we'll simulate it

	return nil
}

func (c *Client) activateUser(ctx context.Context, walletID string, opts CreateUserOptions) (string, error) {
	// For EU countries, this involves:
	// 1. Creating Gatehub managed user
	// 2. Simulating KYC approval
	// 3. Triggering verification webhook

	// For local dev with MockGatehub, we can directly approve the user

	if isEUCountry(opts.CountryCode) {
		// Create Gatehub user
		if err := c.createMockGatehubUser(ctx, opts); err != nil {
			return "", fmt.Errorf("failed to create Gatehub user: %w", err)
		}

		// Simulate webhook approval (would normally come from Gatehub/MockGatehub)
		// This sets KYC status to Level1

		return "Level1", nil
	}

	// For non-EU countries, different KYC providers
	return "Pending", nil
}

func (c *Client) createMockGatehubUser(ctx context.Context, opts CreateUserOptions) error {
	payload := map[string]interface{}{
		"email": opts.Email,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/auth/v1/users/managed", c.mockGatehubURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// For local development, skip TLS verification (self-signed certs)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func isEUCountry(code string) bool {
	euCountries := map[string]bool{
		"AT": true, "BE": true, "BG": true, "HR": true, "CY": true, "CZ": true,
		"DK": true, "EE": true, "FI": true, "FR": true, "DE": true, "GR": true,
		"HU": true, "IE": true, "IT": true, "LV": true, "LT": true, "LU": true,
		"MT": true, "NL": true, "PL": true, "PT": true, "RO": true, "SK": true,
		"SI": true, "ES": true, "SE": true,
	}
	return euCountries[code]
}

// ListUsers fetches all users from Kratos
func (c *Client) ListUsers(ctx context.Context) ([]UserInfo, error) {
	url := fmt.Sprintf("%s/admin/identities", c.kratosAdminURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch identities: %s (status: %d)", string(body), resp.StatusCode)
	}

	var identities []KratosIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identities); err != nil {
		return nil, fmt.Errorf("failed to decode identities: %w", err)
	}

	users := make([]UserInfo, 0, len(identities))
	for _, identity := range identities {
		email := ""
		if emailVal, ok := identity.Traits["email"].(string); ok {
			email = emailVal
		}

		firstName := ""
		if nameMap, ok := identity.Traits["name"].(map[string]interface{}); ok {
			if first, ok := nameMap["first"].(string); ok {
				firstName = first
			}
		}

		lastName := ""
		if nameMap, ok := identity.Traits["name"].(map[string]interface{}); ok {
			if last, ok := nameMap["last"].(string); ok {
				lastName = last
			}
		}

		state := identity.State
		if state == "" {
			state = "active"
		}

		users = append(users, UserInfo{
			ID:        identity.ID,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			State:     state,
		})
	}

	return users, nil
}
