package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
	"github.com/stretchr/testify/assert"
)

// TestSendPersonaAccountTagAdded_Success verifies webhook is sent for account tag added event
func TestSendPersonaAccountTagAdded_Success(t *testing.T) {
	receivedPayload := make(map[string]interface{})
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())
	handler.sendPersonaAccountTagAdded("wallet123", server.URL, "test-secret")

	time.Sleep(100 * time.Millisecond)

	// Verify headers
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	assert.Equal(t, "2023-01-05", receivedHeaders.Get("Persona-Version"))
	assert.NotEmpty(t, receivedHeaders.Get("Persona-Signature"))

	// Verify payload structure
	assert.NotNil(t, receivedPayload["data"])
	data := receivedPayload["data"].(map[string]interface{})
	assert.Equal(t, "event", data["type"])
	assert.NotEmpty(t, data["id"])

	attrs := data["attributes"].(map[string]interface{})
	assert.Equal(t, "account.tag-added", attrs["name"])
	assert.NotEmpty(t, attrs["created-at"])
	assert.NotNil(t, attrs["payload"])
}

// TestSendPersonaAccountTagAdded_SignatureFormat verifies Persona webhook signature format
func TestSendPersonaAccountTagAdded_SignatureFormat(t *testing.T) {
	var receivedSig string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("Persona-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())
	handler.sendPersonaAccountTagAdded("wallet456", server.URL, "persona-test-secret")

	time.Sleep(100 * time.Millisecond)

	// Signature format: "t=timestamp,v1=signature_hex"
	assert.NotEmpty(t, receivedSig)
	assert.True(t, strings.HasPrefix(receivedSig, "t="))
	assert.Contains(t, receivedSig, ",v1=")

	// Verify the signature parts
	parts := strings.Split(receivedSig, ",")
	assert.Len(t, parts, 2)

	// Verify v1 part contains hex characters
	v1Part := parts[1]
	assert.True(t, strings.HasPrefix(v1Part, "v1="))
	hexValue := strings.TrimPrefix(v1Part, "v1=")
	assert.NotEmpty(t, hexValue)

	// Verify it's valid hex (only 0-9, a-f characters)
	for _, ch := range hexValue {
		assert.True(t, (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f'),
			"invalid hex character in signature")
	}
}

// TestSendPersonaAccountTagAdded_PayloadStructure verifies the webhook payload structure
func TestSendPersonaAccountTagAdded_PayloadStructure(t *testing.T) {
	var payload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())
	handler.sendPersonaAccountTagAdded("wallet789", server.URL, "secret")

	time.Sleep(100 * time.Millisecond)

	// Verify top-level data wrapper
	assert.NotNil(t, payload["data"])
	data := payload["data"].(map[string]interface{})

	// Verify event attributes
	attrs := data["attributes"].(map[string]interface{})
	assert.Equal(t, "account.tag-added", attrs["name"])

	// Verify nested account data in payload
	eventPayload := attrs["payload"].(map[string]interface{})
	assert.NotNil(t, eventPayload["data"])

	accountData := eventPayload["data"].(map[string]interface{})
	assert.Equal(t, "account", accountData["type"])
	assert.Equal(t, "wallet789", accountData["id"])

	// Verify account attributes
	accountAttrs := accountData["attributes"].(map[string]interface{})
	assert.Equal(t, "wallet789", accountAttrs["reference-id"])
	assert.NotNil(t, accountAttrs["tags"])

	// Verify tags contain KYC status
	tags := accountAttrs["tags"].([]interface{})
	assert.Len(t, tags, 1)
	assert.Equal(t, "STATUS:KYC-LEVEL:1", tags[0])
}

// TestSendPersonaAccountTagAdded_InvalidURL does not panic on bad URL
func TestSendPersonaAccountTagAdded_InvalidURL(t *testing.T) {
	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())

	// Should not panic or crash with invalid URL
	handler.sendPersonaAccountTagAdded("wallet000", "http://[invalid:url]:9999", "secret")
	// Test passes if no panic
}

// TestSendPersonaAccountTagAdded_ServerErrorHandling does not panic on 5xx response
func TestSendPersonaAccountTagAdded_ServerErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Server Error")
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())

	// Should not panic on server error
	handler.sendPersonaAccountTagAdded("wallet001", server.URL, "secret")
	// Test passes if no panic
}

// TestSendPersonaAccountTagAdded_ClientErrorHandling does not panic on 4xx response
func TestSendPersonaAccountTagAdded_ClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid Request")
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())

	// Should not panic on client error
	handler.sendPersonaAccountTagAdded("wallet002", server.URL, "secret")
	// Test passes if no panic
}

// TestSendPersonaAccountTagAdded_TimestampGeneration verifies unique event IDs
func TestSendPersonaAccountTagAdded_TimestampGeneration(t *testing.T) {
	eventIDs := make(map[string]bool)
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		data := payload["data"].(map[string]interface{})
		eventID := data["id"].(string)
		eventIDs[eventID] = true
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil, testConfig())

	// Send multiple webhooks with microsecond delay
	for i := 0; i < 3; i++ {
		handler.sendPersonaAccountTagAdded(fmt.Sprintf("wallet%d", i), server.URL, "secret")
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(150 * time.Millisecond)

	// Verify unique event IDs were generated
	assert.Equal(t, 3, requestCount)
	assert.Equal(t, 3, len(eventIDs))

	// Verify event IDs follow expected pattern "evt_*"
	for eventID := range eventIDs {
		assert.True(t, strings.HasPrefix(eventID, "evt_"))
	}
}

// TestSendPersonaAccountTagAdded_DifferentWalletsAndSecrets verifies webhook content varies appropriately
func TestSendPersonaAccountTagAdded_DifferentWalletsAndSecrets(t *testing.T) {
	tests := []struct {
		name     string
		walletID string
		secret   string
	}{
		{
			name:     "Wallet A with secret 1",
			walletID: "wallet_a_123",
			secret:   "secret_1",
		},
		{
			name:     "Wallet B with secret 2",
			walletID: "wallet_b_456",
			secret:   "secret_2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPayload map[string]interface{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&receivedPayload)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			store := storage.NewMemoryStorage()
			handler := NewHandler(store, nil, testConfig())
			handler.sendPersonaAccountTagAdded(tt.walletID, server.URL, tt.secret)

			time.Sleep(100 * time.Millisecond)

			// Verify wallet ID appears in payload
			data := receivedPayload["data"].(map[string]interface{})
			attrs := data["attributes"].(map[string]interface{})
			accountData := attrs["payload"].(map[string]interface{})
			accountInfo := accountData["data"].(map[string]interface{})

			assert.Equal(t, tt.walletID, accountInfo["id"])
		})
	}
}
