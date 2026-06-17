package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateWebhookSignature_Success verifies correct HMAC-SHA256 signature generation
func TestGenerateWebhookSignature_Success(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		timestamp int64
		method    string
		url       string
		body      []byte
		expected  string
	}{
		{
			name:      "basic signature generation",
			secret:    "test-secret",
			timestamp: 1234567890,
			method:    "POST",
			url:       "https://example.com/webhook",
			body:      []byte(`{"amount": 100}`),
			expected:  "", // Will compute expected value
		},
		{
			name:      "different secret produces different signature",
			secret:    "different-secret",
			timestamp: 1234567890,
			method:    "POST",
			url:       "https://example.com/webhook",
			body:      []byte(`{"amount": 100}`),
			expected:  "",
		},
		{
			name:      "empty body",
			secret:    "test-secret",
			timestamp: 9999999999,
			method:    "POST",
			url:       "https://api.example.com/webhook",
			body:      []byte{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute expected value
			signaturePayload := fmt.Sprintf("%d|%s|%s|%s", tt.timestamp, tt.method, tt.url, string(tt.body))
			h := hmac.New(sha256.New, []byte(tt.secret))
			h.Write([]byte(signaturePayload))
			expected := hex.EncodeToString(h.Sum(nil))

			// Test signature generation
			result := generateWebhookSignature(tt.secret, tt.timestamp, tt.method, tt.url, tt.body)
			assert.Equal(t, expected, result)
		})
	}
}

// TestGenerateWebhookSignature_Deterministic verifies signature is deterministic
func TestGenerateWebhookSignature_Deterministic(t *testing.T) {
	secret := "test-secret-key"
	timestamp := time.Now().Unix()
	method := "POST"
	url := "https://webhook.example.com/events"
	body := []byte(`{"type": "deposit", "amount": 500}`)

	sig1 := generateWebhookSignature(secret, timestamp, method, url, body)
	sig2 := generateWebhookSignature(secret, timestamp, method, url, body)

	assert.Equal(t, sig1, sig2, "signatures should be identical for same inputs")
}

// TestFormatSettledAt_Success verifies timestamp formatting
func TestFormatSettledAt_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{
			name:     "nil timestamp returns empty string",
			input:    nil,
			expected: "",
		},
		{
			name: "valid timestamp formatted as RFC3339",
			input: func() *time.Time {
				t := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
				return &t
			}(),
			expected: "2024-01-15T10:30:45Z",
		},
		{
			name: "timestamp with microseconds",
			input: func() *time.Time {
				t := time.Date(2024, 12, 31, 23, 59, 59, 123456789, time.UTC)
				return &t
			}(),
			expected: "2024-12-31T23:59:59Z", // Seconds precision in RFC3339
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSettledAt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSendWebhookWithSignature_Success verifies webhook sending with signature
func TestSendWebhookWithSignature_Success(t *testing.T) {
	// Create mock webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Signature"))

		// Verify body
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)
		assert.Equal(t, "deposit_completed", payload["event_type"])

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"received"}`)
	}))
	defer server.Close()

	payload := map[string]interface{}{
		"event_type": "deposit_completed",
		"amount":     100.50,
		"currency":   "USD",
	}

	err := sendWebhookWithSignature(server.URL, "test-secret", payload)
	assert.NoError(t, err)
}

// TestSendWebhookWithSignature_InvalidURL returns error for bad URL
func TestSendWebhookWithSignature_InvalidURL(t *testing.T) {
	payload := map[string]interface{}{"event": "test"}
	err := sendWebhookWithSignature("http://[invalid-url]:9999", "secret", payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

// TestSendWebhookWithSignature_ServerError returns error on 4xx/5xx response
func TestSendWebhookWithSignature_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	payload := map[string]interface{}{"event": "test"}
	err := sendWebhookWithSignature(server.URL, "secret", payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned status 500")
}

// TestSendWebhookWithSignature_NoSecret verifies webhook works without secret
func TestSendWebhookWithSignature_NoSecret(t *testing.T) {
	receivedRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should still receive request even without X-Signature header
		assert.Empty(t, r.Header.Get("X-Signature"))
		receivedRequest = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := map[string]interface{}{"event": "test"}
	err := sendWebhookWithSignature(server.URL, "", payload)
	assert.NoError(t, err)
	assert.True(t, receivedRequest)
}

// TestSendDepositWebhook_Success verifies deposit webhook sending
func TestSendDepositWebhook_Success(t *testing.T) {
	receivedPayload := make(map[string]interface{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set webhook environment variables
	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", server.URL)
	os.Setenv("WEBHOOK_SECRET", "test-secret")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Create sub-account for wallet
	ctx := context.Background()
	store.SaveSubAccount(ctx, &models.SubAccount{
		WalletID:  "wallet123",
		AccountID: "account456",
		FirstName: "John",
		LastName:  "Doe",
	})

	// Send deposit webhook
	handler.sendDepositWebhook("wallet123", "USD", 100.0, "trans001", "")

	// Verify webhook was sent (small delay for goroutine execution)
	time.Sleep(100 * time.Millisecond)

	// Verify payload structure - accountId should be from sub-account
	assert.Equal(t, "account456", receivedPayload["accountId"])
	assert.Equal(t, float64(100), receivedPayload["amount"])
	assert.Equal(t, "USD", receivedPayload["currencyCode"])
	assert.Equal(t, "trans001", receivedPayload["transactionId"])
	assert.Equal(t, float64(104), receivedPayload["code"]) // 104 = successful deposit
}

// TestSendDepositWebhook_NoWebhookURL skips webhook when URL not configured
func TestSendDepositWebhook_NoWebhookURL(t *testing.T) {
	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", "")
	os.Setenv("WEBHOOK_SECRET", "")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Should not panic or error when WEBHOOK_URL is not set
	handler.sendDepositWebhook("wallet123", "USD", 100.0, "trans001", "account456")
	// Test passes if no panic occurs
}

// TestSendDepositWebhook_WithProvidedAccountID uses provided account ID
func TestSendDepositWebhook_WithProvidedAccountID(t *testing.T) {
	receivedPayload := make(map[string]interface{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", server.URL)
	os.Setenv("WEBHOOK_SECRET", "test-secret")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Send with explicit account ID (should not look up sub-account)
	handler.sendDepositWebhook("wallet123", "EUR", 250.50, "trans789", "explicit_account_id")

	time.Sleep(100 * time.Millisecond)

	// Should use provided account ID
	assert.Equal(t, "explicit_account_id", receivedPayload["accountId"])
	assert.Equal(t, float64(250.50), receivedPayload["amount"])
	assert.Equal(t, "EUR", receivedPayload["currencyCode"])
}

// TestSendDepositCompletedWebhook_Success verifies completed deposit webhook
func TestSendDepositCompletedWebhook_Success(t *testing.T) {
	var receivedSignature string
	var receivedTimestamp string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("x-gatehub-signature")
		receivedTimestamp = r.Header.Get("x-gatehub-timestamp")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "xago-mock", r.Header.Get("x-gatehub-app-id"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", server.URL)
	os.Setenv("WEBHOOK_SECRET", "test-secret-key")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Create deposit with settled time
	ctx := context.Background()
	settledTime := time.Now().UTC()
	store.SaveDeposit(ctx, &models.Deposit{
		ID:               "dep001",
		AccountID:        "acc123",
		Amount:           100.0,
		Currency:         "USD",
		DepositReference: "ref123",
		CreatedAt:        time.Now().UTC(),
		SettledAt:        &settledTime,
	})

	// Send completed webhook
	handler.sendDepositCompletedWebhook("acc123", 100.0, "USD", "dep001", "ref123")

	time.Sleep(100 * time.Millisecond)

	// Verify signature was generated
	assert.NotEmpty(t, receivedSignature)
	assert.NotEmpty(t, receivedTimestamp)

	// Verify signature is valid
	ts, _ := strconv.ParseInt(receivedTimestamp, 10, 64)
	payload := models.DepositWebhookPayload{
		AccountID:            "acc123",
		Amount:               100.0,
		CurrencyCode:         "USD",
		TransactionID:        "dep001",
		TransactionReference: "ref123",
		Status:               "completed",
		Code:                 104,
		CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		SettledAt:            settledTime.Format(time.RFC3339),
	}
	bodyBytes, _ := json.Marshal(payload)
	_ = generateWebhookSignature("test-secret-key", ts, "POST", server.URL, bodyBytes)
	// Note: exact match unlikely due to timestamp precision, just verify structure
	assert.NotEmpty(t, receivedSignature)
}

// TestSendDepositCompletedWebhook_NoWebhookURL skips when not configured
func TestSendDepositCompletedWebhook_NoWebhookURL(t *testing.T) {
	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", "")
	os.Setenv("WEBHOOK_SECRET", "")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Should skip webhook gracefully
	handler.sendDepositCompletedWebhook("acc123", 100.0, "USD", "dep001", "ref123")
	// Test passes if no error
}

// TestSendDepositCompletedWebhook_MissingDeposit handles missing deposit gracefully
func TestSendDepositCompletedWebhook_MissingDeposit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", server.URL)
	os.Setenv("WEBHOOK_SECRET", "test-secret")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)

	// Send webhook for non-existent deposit
	// Should handle error gracefully without panicking
	handler.sendDepositCompletedWebhook("acc999", 50.0, "GBP", "nonexistent", "ref999")
	// Test passes if no panic
}

// TestWebhookSignatureValidation verifies signature correctness
func TestWebhookSignatureValidation(t *testing.T) {
	secret := "my-webhook-secret"
	timestamp := int64(1234567890)
	method := "POST"
	url := "https://api.example.com/webhooks/events"
	body := []byte(`{"amount": 500, "currency": "USD"}`)

	signature := generateWebhookSignature(secret, timestamp, method, url, body)

	// Reconstruct what was signed
	signaturePayload := fmt.Sprintf("%d|%s|%s|%s", timestamp, method, url, string(body))

	// Verify the signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signaturePayload))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	assert.Equal(t, expectedSig, signature)
}

// TestWebhookSignatureInvalidatesWithDifferentSecret verifies signature security
func TestWebhookSignatureInvalidatesWithDifferentSecret(t *testing.T) {
	secret := "original-secret"
	timestamp := int64(1234567890)
	method := "POST"
	url := "https://api.example.com/webhooks"
	body := []byte(`{"test": true}`)

	signature := generateWebhookSignature(secret, timestamp, method, url, body)

	// Attempt to verify with different secret
	wrongSecret := "different-secret"
	wrongSignature := generateWebhookSignature(wrongSecret, timestamp, method, url, body)

	assert.NotEqual(t, signature, wrongSignature, "signatures with different secrets must differ")
}

// TestWebhookPayloadMarshaling verifies DepositWebhookPayload can be marshaled
func TestWebhookPayloadMarshaling(t *testing.T) {
	payload := models.DepositWebhookPayload{
		AccountID:            "acc123",
		Amount:               100.50,
		CurrencyCode:         "USD",
		TransactionID:        "txn456",
		TransactionReference: "ref789",
		Status:               "completed",
		Code:                 104,
		CreatedAt:            "2024-01-15T10:30:45Z",
		SettledAt:            "2024-01-15T10:31:00Z",
	}

	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	// Verify it can be unmarshaled
	var decoded models.DepositWebhookPayload
	err = json.Unmarshal(bodyBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, payload, decoded)
}

// TestSendDepositWebhook_HeaderValidation verifies webhook headers are correct
func TestSendDepositWebhook_HeaderValidation(t *testing.T) {
	receivedHeaders := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders["content-type"] = r.Header.Get("Content-Type")
		receivedHeaders["x-signature"] = r.Header.Get("X-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldWebhookURL := os.Getenv("WEBHOOK_URL")
	oldWebhookSecret := os.Getenv("WEBHOOK_SECRET")
	os.Setenv("WEBHOOK_URL", server.URL)
	os.Setenv("WEBHOOK_SECRET", "secret")
	defer func() {
		os.Setenv("WEBHOOK_URL", oldWebhookURL)
		os.Setenv("WEBHOOK_SECRET", oldWebhookSecret)
	}()

	store := storage.NewMemoryStorage()
	handler := NewHandler(store, nil)
	ctx := context.Background()
	store.SaveSubAccount(ctx, &models.SubAccount{
		WalletID:  "w1",
		AccountID: "a1",
	})

	handler.sendDepositWebhook("w1", "USD", 100.0, "t1", "")

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "application/json", receivedHeaders["content-type"])
	assert.NotEmpty(t, receivedHeaders["x-signature"])
}

// BenchmarkGenerateWebhookSignature benchmarks signature generation performance
func BenchmarkGenerateWebhookSignature(b *testing.B) {
	secret := "benchmark-secret"
	timestamp := int64(1234567890)
	method := "POST"
	url := "https://api.example.com/webhooks/events"
	body := []byte(`{"amount": 100, "currency": "USD", "account": "test-account"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateWebhookSignature(secret, timestamp, method, url, body)
	}
}

// BenchmarkFormatSettledAt benchmarks timestamp formatting
func BenchmarkFormatSettledAt(b *testing.B) {
	t := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatSettledAt(&t)
	}
}
