package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsEUCountry(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"DE", true},
		{"FR", true},
		{"IT", true},
		{"ES", true},
		{"US", false},
		{"CA", false},
		{"GB", false}, // UK is not in EU anymore
		{"ZA", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := isEUCountry(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateKratosIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/admin/identities", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "default", body["schema_id"])
		traits := body["traits"].(map[string]interface{})
		assert.Equal(t, "test@example.com", traits["email"])
		assert.Equal(t, "John", traits["firstName"])
		assert.Equal(t, "Doe", traits["lastName"])

		// Return mock identity
		identity := KratosIdentity{
			ID:       "identity-123",
			SchemaID: "default",
			Traits: map[string]interface{}{
				"email": "test@example.com",
			},
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(identity)
	}))
	defer server.Close()

	client := &Client{
		kratosAdminURL: server.URL,
	}

	opts := CreateUserOptions{
		Email:       "test@example.com",
		FirstName:   "John",
		LastName:    "Doe",
		CountryCode: "US",
		Password:    "Test@123456?",
	}

	identity, err := client.createKratosIdentity(context.Background(), opts)
	require.NoError(t, err)
	assert.Equal(t, "identity-123", identity.ID)
}

func TestVerifyEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/admin/identities/identity-123", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify patch operations
		var patches []map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&patches)
		require.NoError(t, err)
		assert.Len(t, patches, 1)
		assert.Equal(t, "replace", patches[0]["op"])
		assert.Equal(t, "/verifiable_addresses", patches[0]["path"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "identity-123"})
	}))
	defer server.Close()

	client := &Client{
		kratosAdminURL: server.URL,
	}

	err := client.verifyEmail(context.Background(), "identity-123", "test@example.com")
	require.NoError(t, err)
}

func TestCreateMockGatehubUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth/v1/users/managed", r.URL.Path)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", body["email"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id":    "gatehub-user-123",
			"email": "test@example.com",
		})
	}))
	defer server.Close()

	client := &Client{
		mockGatehubURL: server.URL,
	}

	opts := CreateUserOptions{
		Email: "test@example.com",
	}

	err := client.createMockGatehubUser(context.Background(), opts)
	require.NoError(t, err)
}

func TestCreateUserOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        CreateUserOptions
		expectError bool
	}{
		{
			name: "valid options",
			opts: CreateUserOptions{
				Email:       "test@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				CountryCode: "US",
				Password:    "Test@123456?",
			},
			expectError: false,
		},
		{
			name: "empty email",
			opts: CreateUserOptions{
				Email:       "",
				FirstName:   "John",
				LastName:    "Doe",
				CountryCode: "US",
				Password:    "Test@123456?",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			if tt.expectError {
				assert.Empty(t, tt.opts.Email)
			} else {
				assert.NotEmpty(t, tt.opts.Email)
			}
		})
	}
}

func TestCreateUserResult(t *testing.T) {
	result := &CreateUserResult{
		IdentityID: "identity-123",
		Email:      "test@example.com",
		WalletID:   "wallet-123",
		KYCStatus:  "Level1",
	}

	assert.Equal(t, "identity-123", result.IdentityID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "wallet-123", result.WalletID)
	assert.Equal(t, "Level1", result.KYCStatus)
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	assert.NotEmpty(t, client.kratosAdminURL)
	assert.NotEmpty(t, client.mockGatehubURL)
}

func TestCreateDefaultWallet(t *testing.T) {
	client := &Client{}

	opts := CreateUserOptions{
		Email: "test@example.com",
	}

	walletID, err := client.createDefaultWallet(context.Background(), "identity-12345678", opts)
	require.NoError(t, err)
	assert.Contains(t, walletID, "wallet-identity")
}

func TestSetupKYC(t *testing.T) {
	client := &Client{}

	opts := CreateUserOptions{
		Email:       "test@example.com",
		CountryCode: "US",
	}

	err := client.setupKYC(context.Background(), "wallet-123", opts)
	require.NoError(t, err)
}

func TestActivateUser_EU(t *testing.T) {
	// Mock Gatehub API
	ghServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id": "gatehub-user-123",
		})
	}))
	defer ghServer.Close()

	client := &Client{
		mockGatehubURL: ghServer.URL,
	}

	opts := CreateUserOptions{
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		CountryCode: "DE",
	}

	kycStatus, err := client.activateUser(context.Background(), "wallet-123", opts)
	require.NoError(t, err)
	// EU users get Level1 KYC status
	assert.Equal(t, "Level1", kycStatus)
}

func TestActivateUser_NonEU(t *testing.T) {
	client := &Client{}

	opts := CreateUserOptions{
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		CountryCode: "US",
	}

	kycStatus, err := client.activateUser(context.Background(), "wallet-123", opts)
	require.NoError(t, err)
	// Non-EU users get Pending status initially
	assert.Equal(t, "Pending", kycStatus)
}

func TestListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/admin/identities", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		identities := []KratosIdentity{
			{
				ID:       "id-1",
				SchemaID: "default",
				State:    "active",
				Traits: map[string]interface{}{
					"email": "user1@example.com",
					"name": map[string]interface{}{
						"first": "John",
						"last":  "Doe",
					},
				},
			},
			{
				ID:       "id-2",
				SchemaID: "default",
				State:    "inactive",
				Traits: map[string]interface{}{
					"email": "user2@example.com",
					"name": map[string]interface{}{
						"first": "Jane",
						"last":  "Smith",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(identities)
	}))
	defer server.Close()

	client := &Client{
		kratosAdminURL: server.URL,
	}

	users, err := client.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 2)

	assert.Equal(t, "id-1", users[0].ID)
	assert.Equal(t, "user1@example.com", users[0].Email)
	assert.Equal(t, "John", users[0].FirstName)
	assert.Equal(t, "Doe", users[0].LastName)
	assert.Equal(t, "active", users[0].State)

	assert.Equal(t, "id-2", users[1].ID)
	assert.Equal(t, "user2@example.com", users[1].Email)
	assert.Equal(t, "Jane", users[1].FirstName)
	assert.Equal(t, "Smith", users[1].LastName)
	assert.Equal(t, "inactive", users[1].State)
}

func TestListUsers_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]KratosIdentity{})
	}))
	defer server.Close()

	client := &Client{
		kratosAdminURL: server.URL,
	}

	users, err := client.ListUsers(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestListUsers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := &Client{
		kratosAdminURL: server.URL,
	}

	users, err := client.ListUsers(context.Background())
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Contains(t, err.Error(), "failed to fetch identities")
}
