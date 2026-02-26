package totp

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Client struct {
	kratosAdminURL string
	secretsFile    string
}

type KratosIdentity struct {
	ID          string                 `json:"id"`
	Traits      map[string]interface{} `json:"traits"`
	Credentials map[string]interface{} `json:"credentials"`
}

type StoredSecret struct {
	IdentityID string    `json:"identity_id"`
	Email      string    `json:"email"`
	Secret     string    `json:"secret"`
	CreatedAt  time.Time `json:"created_at"`
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

	// Store TOTP secrets in a local file for persistence
	// This is a workaround since Kratos admin API doesn't easily support direct credential updates
	secretsFile := filepath.Join(os.TempDir(), "local-dev-tool-totp-secrets.json")

	return &Client{
		kratosAdminURL: getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		secretsFile:    secretsFile,
	}
}

var ErrTOTPNotConfigured = fmt.Errorf("TOTP not configured")

func (c *Client) GenerateCode(ctx context.Context, email string) (string, error) {
	// 1. Find user by email
	identity, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	// 2. Get TOTP secret from Kratos or local storage
	secret, err := c.getTOTPSecret(ctx, identity)
	if err != nil {
		return "", err
	}

	// 3. Generate TOTP code
	code, err := generateTOTPCode(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	return code, nil
}

func (c *Client) StoreSecret(ctx context.Context, email, secret string) error {
	// Find user by email
	identity, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Validate the secret by trying to generate a code
	_, err = generateTOTPCode(secret)
	if err != nil {
		return fmt.Errorf("invalid TOTP secret: %w", err)
	}

	// Store the secret
	return c.storeSecret(identity.ID, email, secret)
}

func (c *Client) getUserByEmail(ctx context.Context, email string) (*KratosIdentity, error) {
	// Get all identities and filter by email
	req, err := http.NewRequestWithContext(ctx, "GET", c.kratosAdminURL+"/admin/identities", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var identities []KratosIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identities); err != nil {
		return nil, err
	}

	// Find user by email
	for _, identity := range identities {
		if traits, ok := identity.Traits["email"].(string); ok && traits == email {
			// Fetch full identity details to get credentials
			return c.getIdentityByID(ctx, identity.ID)
		}
	}

	return nil, fmt.Errorf("user with email '%s' not found", email)
}

func (c *Client) getIdentityByID(ctx context.Context, id string) (*KratosIdentity, error) {
	// Include credentials to get TOTP config
	url := fmt.Sprintf("%s/admin/identities/%s?include_credential=totp", c.kratosAdminURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var identity KratosIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

func (c *Client) getTOTPSecret(ctx context.Context, identity *KratosIdentity) (string, error) {
	// First check if TOTP is configured in Kratos
	secret, err := c.extractTOTPSecret(identity)
	if err == nil {
		// TOTP configured in Kratos
		return secret, nil
	}

	// Check local storage for previously entered secret
	storedSecret, found := c.getStoredSecret(identity.ID)
	if found {
		// Secret exists in local storage
		return storedSecret, nil
	}

	// TOTP not configured anywhere
	return "", ErrTOTPNotConfigured
}

func (c *Client) extractTOTPSecret(identity *KratosIdentity) (string, error) {
	if identity.Credentials == nil {
		return "", fmt.Errorf("user has no TOTP configured")
	}

	totpCred, ok := identity.Credentials["totp"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("user has no TOTP configured")
	}

	config, ok := totpCred["config"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid TOTP configuration")
	}

	secret, ok := config["totp_url"].(string)
	if !ok {
		return "", fmt.Errorf("TOTP secret not found")
	}

	// Extract secret from otpauth:// URL
	// Format: otpauth://totp/interledger.app:email?secret=SECRET&issuer=interledger.app
	return extractSecretFromURL(secret)
}

func (c *Client) getStoredSecret(identityID string) (string, bool) {
	secrets := c.loadSecrets()
	for _, stored := range secrets {
		if stored.IdentityID == identityID {
			return stored.Secret, true
		}
	}
	return "", false
}

func (c *Client) storeSecret(identityID, email, secret string) error {
	secrets := c.loadSecrets()

	// Check if already exists and update
	found := false
	for i, stored := range secrets {
		if stored.IdentityID == identityID {
			secrets[i].Secret = secret
			secrets[i].CreatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		secrets = append(secrets, StoredSecret{
			IdentityID: identityID,
			Email:      email,
			Secret:     secret,
			CreatedAt:  time.Now(),
		})
	}

	return c.saveSecrets(secrets)
}

func (c *Client) loadSecrets() []StoredSecret {
	data, err := os.ReadFile(c.secretsFile)
	if err != nil {
		// File doesn't exist or can't be read, return empty
		return []StoredSecret{}
	}

	var secrets []StoredSecret
	if err := json.Unmarshal(data, &secrets); err != nil {
		// Invalid JSON, return empty
		return []StoredSecret{}
	}

	return secrets
}

func (c *Client) saveSecrets(secrets []StoredSecret) error {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.secretsFile, data, 0600)
}



func extractSecretFromURL(totpURL string) (string, error) {
	// Parse the otpauth:// URL to extract the secret parameter
	// Format: otpauth://totp/interledger.app:email?secret=SECRET&issuer=interledger.app

	// Find the secret= parameter
	secretPrefix := "secret="
	secretStart := -1
	for i := 0; i < len(totpURL); i++ {
		if i+len(secretPrefix) <= len(totpURL) && totpURL[i:i+len(secretPrefix)] == secretPrefix {
			secretStart = i + len(secretPrefix)
			break
		}
	}

	if secretStart == -1 {
		return "", fmt.Errorf("secret parameter not found in TOTP URL")
	}

	// Find the end of the secret (next & or end of string)
	secretEnd := secretStart
	for secretEnd < len(totpURL) && totpURL[secretEnd] != '&' {
		secretEnd++
	}

	return totpURL[secretStart:secretEnd], nil
}

func generateTOTPCode(secret string) (string, error) {
	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Get current time counter (30-second intervals)
	counter := time.Now().Unix() / 30

	// Convert counter to byte array
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate 6-digit code
	code := truncated % 1000000

	return fmt.Sprintf("%06d", code), nil
}
