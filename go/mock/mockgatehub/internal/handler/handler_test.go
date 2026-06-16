package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateMaskedPan verifies masked PAN format is numeric-only with correct structure
func TestGenerateMaskedPan(t *testing.T) {
	panPattern := regexp.MustCompile(`^\d{6}\*{6}\d{4}$`)

	for i := 0; i < 100; i++ {
		pan := generateMaskedPan()
		assert.Regexp(t, panPattern, pan, "masked PAN should match dddddd******dddd format, got: %s", pan)
	}
}

// TestRandomDigits verifies randomDigits returns only decimal digits of correct length
func TestRandomDigits(t *testing.T) {
	for _, n := range []int{1, 4, 6, 10} {
		digits := randomDigits(n)
		assert.Len(t, digits, n)
		assert.Regexp(t, regexp.MustCompile(`^\d+$`), digits, "randomDigits(%d) should be all numeric, got: %s", n, digits)
	}
}

// TestCreateTransactionExternalDeposit verifies external deposits are created and return correct status
func TestCreateTransactionExternalDeposit(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "") // No URL - won't send webhooks
	handler := NewHandler(store, webhookManager)

	// Create transaction request for external deposit
	body := map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         1, // External deposit
		"deposit_type": "external",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Verify transaction fields
	assert.NotEmpty(t, response["uuid"])
	assert.Equal(t, "100.00", response["amount"])
	assert.Equal(t, "USD", response["currency"])
	assert.Equal(t, "external", response["deposit_type"])
	assert.Equal(t, float64(1), response["status"]) // status=1 means completed

	// Verify balance was updated (note: test user may have pre-seeded balance, so just verify it increased)
	balance, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Greater(t, balance, 0.0, "Balance should be positive after deposit")
}

// TestCreateTransactionHostedDeposit verifies hosted deposits are created successfully
// This is the critical fix - hosted transfers (type=2) must send webhooks to prevent
// PayIn workflow from hanging indefinitely waiting for webhook or 20-minute polling
func TestCreateTransactionHostedDeposit(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "") // No URL - won't send webhooks
	handler := NewHandler(store, webhookManager)

	// Create transaction request for hosted deposit (type=2)
	body := map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       50.00,
		"currency":     "EUR",
		"type":         2, // Hosted transfer
		"deposit_type": "hosted",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected successful transaction creation for hosted deposit")

	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// CRITICAL FIX: Verify hosted transfer was created successfully
	// (Previously this would fail or not trigger webhooks, causing PayIn workflow to hang indefinitely)
	assert.NotEmpty(t, response["uuid"], "Transaction UUID should not be empty")
	assert.Equal(t, "50.00", response["amount"])
	assert.Equal(t, "EUR", response["currency"])
	assert.Equal(t, "hosted", response["deposit_type"])
	assert.Equal(t, float64(1), response["status"]) // status=1 means completed

	// Verify balance was updated
	balance, _ := store.GetBalance(consts.TestUser1ID, "EUR")
	assert.Greater(t, balance, 0.0, "Balance should be positive after deposit")
}

// TestCreateTransactionMissingUserID verifies validation
func TestCreateTransactionMissingUserID(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Missing user_id
	body := map[string]interface{}{
		"amount":   100.00,
		"currency": "USD",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	// Should fail with bad request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateTransactionMultipleCurrencies verifies different currencies can be handled
func TestCreateTransactionMultipleCurrencies(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	currencies := []string{"USD", "EUR", "GBP", "XRP"}
	for _, curr := range currencies {
		body := map[string]interface{}{
			"user_id":  consts.TestUser1ID,
			"amount":   25.00,
			"currency": curr,
			"type":     2, // Hosted
		}

		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.CreateTransaction(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code, "Failed to create transaction for %s", curr)

		var response map[string]interface{}
		_ = json.NewDecoder(rr.Body).Decode(&response)
		assert.Equal(t, curr, response["currency"])
	}

	// Verify all balances were updated
	for _, curr := range currencies {
		balance, _ := store.GetBalance(consts.TestUser1ID, curr)
		assert.Greater(t, balance, 0.0, "Balance should be positive for %s after deposit", curr)
	}
}

// TestTransactionCompleteHandlerMissingAmount verifies that missing amount returns 400
func TestTransactionCompleteHandlerMissingAmount(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Request body missing amount field
	body := map[string]interface{}{
		"currency": "USD",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should fail with bad request - amount is required
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request when amount is missing")
}

// TestTransactionCompleteHandlerMissingCurrency verifies that missing currency returns 400
func TestTransactionCompleteHandlerMissingCurrency(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Request body missing currency field
	body := map[string]interface{}{
		"amount": "100.00",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should fail with bad request - currency is required
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request when currency is missing")
}

// TestTransactionCompleteHandlerEmptyBody verifies that empty body returns 400
func TestTransactionCompleteHandlerEmptyBody(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Empty body
	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should fail with bad request - both amount and currency are missing
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request when body is empty")
}

// TestTransactionCompleteHandlerInvalidAmount verifies that non-numeric amount returns 400
func TestTransactionCompleteHandlerInvalidAmount(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Invalid amount (not a number)
	body := map[string]interface{}{
		"amount":   "not-a-number",
		"currency": "USD",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should fail with bad request - amount is invalid
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request when amount is not numeric")
}

// TestTransactionCompleteHandlerInvalidCurrency verifies that unknown currency returns 400
func TestTransactionCompleteHandlerInvalidCurrency(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Invalid currency
	body := map[string]interface{}{
		"amount":   "100.00",
		"currency": "UNKNOWN",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should fail with bad request - currency is unknown
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 Bad Request when currency is unknown")
}

// TestTransactionCompleteHandlerValid verifies that valid request succeeds
func TestTransactionCompleteHandlerValid(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	handler := NewHandler(store, webhookManager)

	// Create a wallet first
	createTestWallet(t, store, consts.TestUser1ID)

	// Valid request
	body := map[string]interface{}{
		"amount":   "100.00",
		"currency": "USD",
	}

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Map token to user
	handler.tokenToUser.Store("test-token", consts.TestUser1ID)

	rr := httptest.NewRecorder()
	handler.TransactionCompleteHandler(rr, req)

	// Should succeed
	assert.Equal(t, http.StatusOK, rr.Code, "Expected 200 OK for valid transaction complete request")

	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

// TestRequestLoggerMultiValueHeaders verifies that multi-value headers are properly joined
func TestRequestLoggerMultiValueHeaders(t *testing.T) {
	store := storage.NewMemoryStorage()
	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	h := NewHandler(store, webhookManager)

	// Create a simple test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap it with the RequestLogger middleware
	loggedHandler := h.RequestLogger(testHandler)

	// Create a request with multi-value headers
	req := httptest.NewRequest("GET", "/test", bytes.NewReader([]byte("{}")))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("X-Custom", "value1")
	req.Header.Add("X-Custom", "value2")
	req.Header.Add("X-Custom", "value3")

	// Make the request
	rr := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rr, req)

	// Verify the response is OK (middleware should not interfere with request processing)
	assert.Equal(t, http.StatusOK, rr.Code, "RequestLogger middleware should not affect request processing")
}

// Helper function to create a test wallet
func createTestWallet(t *testing.T, store storage.Storage, userID string) {
	wallet := &models.Wallet{
		UserID:    userID,
		Name:      "Test Wallet",
		Address:   "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
		Type:      0,
		Network:   30, // XRP Ledger
		CreatedAt: time.Now(),
	}
	err := store.CreateWallet(wallet)
	require.NoError(t, err)
}

// setupSCATestHandler creates a handler with a test user, wallet, balance, and optional org config.
func setupSCATestHandler(t *testing.T, orgAPIBaseURL string) (*Handler, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)

	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	h := NewHandler(store, wm)

	createTestWallet(t, store, consts.TestUser1ID)
	h.tokenToUser.Store("sca-token", consts.TestUser1ID)

	if orgAPIBaseURL != "" {
		// Seeder already creates "default-org"; update it with the provided apiBaseUrl
		org, err := store.GetOrganization("default-org")
		require.NoError(t, err)
		org.APIBaseURL = orgAPIBaseURL
		org.TwoFAType = "totp"
		require.NoError(t, store.UpdateOrganization(org))
	}

	return h, store
}

func postTransaction(h *Handler, paymentType string, body map[string]interface{}) *httptest.ResponseRecorder {
	bodyBytes, _ := json.Marshal(body)
	url := fmt.Sprintf("/transaction/complete?paymentType=%s&bearer=sca-token", paymentType)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.TransactionCompleteHandler(w, req)
	return w
}

func TestTransactionSCA_DepositSuccess(t *testing.T) {
	var receivedBody map[string]interface{}
	var receivedPath string
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, store := setupSCATestHandler(t, mock2FA.URL)

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "123456",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fmt.Sprintf("/v1/users/managed/%s/2fa", consts.TestUser1ID), receivedPath)
	assert.Equal(t, "VERIFY", receivedBody["action"])
	assert.Equal(t, "123456", receivedBody["code"])

	// Balance should have increased
	balance, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Greater(t, balance, 10000.0) // seeded with 10,000
}

func TestTransactionSCA_DepositRejected(t *testing.T) {
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}))
	defer mock2FA.Close()

	h, store := setupSCATestHandler(t, mock2FA.URL)

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "wrong",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Balance should be unchanged
	balance, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Equal(t, 10000.0, balance)
}

func TestTransactionSCA_NoOrgConfig(t *testing.T) {
	h, _ := setupSCATestHandler(t, "") // no org

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "123456",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["message"], "no organization callback URL configured")
}

func TestTransactionSCA_InvalidBearer(t *testing.T) {
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, _ := setupSCATestHandler(t, mock2FA.URL)

	// Use an unknown bearer token
	bodyBytes, _ := json.Marshal(map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "123456",
	})
	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=deposit&bearer=unknown-token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.TransactionCompleteHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionSCA_NetworkError(t *testing.T) {
	h, _ := setupSCATestHandler(t, "https://127.0.0.1:1") // unroutable

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "123456",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionSCA_WithdrawalSuccess(t *testing.T) {
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, store := setupSCATestHandler(t, mock2FA.URL)

	w := postTransaction(h, "withdrawal", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "654321",
	})

	assert.Equal(t, http.StatusOK, w.Code)

	// Balance is not deducted until the withdrawal event is triggered via the admin endpoint
	balance, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Equal(t, 10000.0, balance)
}

func TestTransactionSCA_SkippedWhenNotTriggered(t *testing.T) {
	// No SCA fields — should succeed without any 2FA callback, even with no org config
	h, _ := setupSCATestHandler(t, "")

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":   "50.00",
		"currency": "USD",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTransactionSCA_IntegratorReturns500(t *testing.T) {
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mock2FA.Close()

	h, _ := setupSCATestHandler(t, mock2FA.URL)

	w := postTransaction(h, "deposit", map[string]interface{}{
		"amount":      "50.00",
		"currency":    "USD",
		"trigger_sca": true,
		"totp_code":   "123456",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── HealthCheck ──

func TestHealthCheck(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.HealthCheck(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.Contains(t, rec.Body.String(), `"service":"gitlab.com/fynbos/mock/mockgatehub"`)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

// ── extractBearerFromAuthHeader ──

func TestExtractBearerFromAuthHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid", "Bearer abc123", "abc123"},
		{"empty", "", ""},
		{"too_short", "Bear", ""},
		{"no_bearer_prefix", "Basic abc123", ""},
		{"bearer_only", "Bearer ", ""},
		{"with_spaces", "Bearer tok en", "tok en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractBearerFromAuthHeader(tt.header))
		})
	}
}

// ── validateTransactionRequest ──

func TestValidateTransactionRequest(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	tests := []struct {
		name    string
		req     *TransactionRequest
		wantErr bool
	}{
		{"valid", &TransactionRequest{Amount: "100.00", Currency: "USD"}, false},
		{"missing_amount", &TransactionRequest{Currency: "USD"}, true},
		{"missing_currency", &TransactionRequest{Amount: "100.00"}, true},
		{"invalid_amount", &TransactionRequest{Amount: "abc", Currency: "USD"}, true},
		{"unsupported_currency", &TransactionRequest{Amount: "10", Currency: "FAKE"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.validateTransactionRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── isSupportedCurrency ──

func TestIsSupportedCurrency(t *testing.T) {
	assert.True(t, isSupportedCurrency("USD"))
	assert.True(t, isSupportedCurrency("EUR"))
	assert.True(t, isSupportedCurrency("XRP"))
	assert.False(t, isSupportedCurrency("FAKE"))
	assert.False(t, isSupportedCurrency(""))
}

// ── RequestLogger preserves body ──

func TestRequestLoggerPreservesBody(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	var capturedBody string
	inner := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		rw.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"key":"value"}`))
	rec := httptest.NewRecorder()
	h.RequestLogger(inner).ServeHTTP(rec, req)

	assert.Equal(t, `{"key":"value"}`, capturedBody)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── TransactionCompleteHandler OPTIONS ──

func TestTransactionCompleteHandler_CORS_Options(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodOptions, "/transaction/complete?paymentType=deposit", nil)
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

// ── extractUserFromBearer ──

func TestExtractUserFromBearer(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	h.tokenToUser.Store("known-tok", "user-123")

	assert.Equal(t, "user-123", h.extractUserFromBearer("known-tok"))
	assert.Equal(t, "", h.extractUserFromBearer("unknown-really-long-token-value"))
}

// ── processWithdrawal insufficient balance ──

func TestProcessWithdrawal_InsufficientBalance(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	createTestWallet(t, store, consts.TestUser1ID)
	h.tokenToUser.Store("bearer-insuff", consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=withdrawal&bearer=bearer-insuff",
		strings.NewReader(`{"amount":"999999.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Insufficient balance")
}

// ── processWithdrawal no wallets ──

func TestProcessWithdrawal_NoWallets(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	h.tokenToUser.Store("bearer-nowal", consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=withdrawal&bearer=bearer-nowal",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no wallets")
}

// ── min helper ──

func TestMinHelper(t *testing.T) {
	assert.Equal(t, 3, min(3, 5))
	assert.Equal(t, 3, min(5, 3))
	assert.Equal(t, 0, min(0, 0))
}

// ── RootHandler ──

func TestRootHandler_MissingBearer(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.RootHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing bearer token")
}

// ── TransactionCompleteHandler with Authorization header ──

func TestTransactionCompleteHandler_AuthHeader(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	createTestWallet(t, store, consts.TestUser1ID)
	h.tokenToUser.Store("auth-bearer-tok", consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=deposit",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer auth-bearer-tok")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── TransactionCompleteHandler missing bearer ──

func TestTransactionCompleteHandler_MissingBearer(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=deposit",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Missing bearer token")
}

// ── processDeposit invalid bearer ──

func TestProcessDeposit_InvalidBearer(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=deposit&bearer=unknown-token-value-long",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid bearer token")
}

// ── processWithdrawal invalid bearer ──

func TestProcessWithdrawal_InvalidBearer(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=withdrawal&bearer=unknown-token-value-long",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ── processDeposit auto-creates wallet ──

func TestProcessDeposit_AutoCreatesWallet(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	h.tokenToUser.Store("auto-wal-tok", consts.TestUser1ID)

	// Don't create a wallet — processDeposit should auto-create one
	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=deposit&bearer=auto-wal-tok",
		strings.NewReader(`{"amount":"25.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Verify balance increased
	bal, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.InDelta(t, 10025.0, bal, 0.01)
}

// ── processWithdrawal success ──

func TestProcessWithdrawal_Success(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	createTestWallet(t, store, consts.TestUser1ID)
	h.tokenToUser.Store("wd-tok", consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=withdrawal&bearer=wd-tok",
		strings.NewReader(`{"amount":"50.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "pending")
	// Balance is not deducted until the withdrawal event is triggered via the admin endpoint
	bal, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Equal(t, 10000.0, bal)
}

// ── generic paymentType (not deposit/withdrawal) ──

func TestTransactionComplete_GenericType(t *testing.T) {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "s", nil, store, "")
	h := NewHandler(store, wm)

	h.tokenToUser.Store("gen-tok", "user-1")

	req := httptest.NewRequest(http.MethodPost, "/transaction/complete?paymentType=exchange&bearer=gen-tok",
		strings.NewReader(`{"amount":"10.00","currency":"USD"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TransactionCompleteHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Transaction completed")
}

// ── withdrawal helpers ──

func createPendingWithdrawal(t *testing.T, store storage.Storage, userID, currency, amount string) string {
	t.Helper()
	tx := &models.Transaction{
		UserID:      userID,
		Amount:      amount,
		TotalAmount: amount,
		Fee:         "0.00",
		Currency:    currency,
		DepositType: consts.DepositTypeWithdrawal,
		Type:        consts.TransactionTypeWithdrawal,
		Status:      consts.TransactionStatusPending,
	}
	require.NoError(t, store.CreateTransaction(tx))
	return tx.ID
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── ListWithdrawals ──

func TestListWithdrawals_NoPendingWithdrawals(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+consts.TestUser1ID+"/withdrawals?status=pending", nil)
	req = withURLParam(req, "userID", consts.TestUser1ID)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Empty(t, result)
}

func TestListWithdrawals_ReturnsPendingWithdrawals(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+consts.TestUser1ID+"/withdrawals?status=pending", nil)
	req = withURLParam(req, "userID", consts.TestUser1ID)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Len(t, result, 1)
	assert.Equal(t, "50.00", result[0]["amount"])
}

func TestListWithdrawals_ExcludesCompletedWhenFilteringPending(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")
	require.NoError(t, store.UpdateTransactionStatus(txID, consts.TransactionStatusCompleted))

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+consts.TestUser1ID+"/withdrawals?status=pending", nil)
	req = withURLParam(req, "userID", consts.TestUser1ID)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Empty(t, result)
}

func TestListWithdrawals_NoStatusFilterReturnsAll(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")
	require.NoError(t, store.UpdateTransactionStatus(txID, consts.TransactionStatusCompleted))
	createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "30.00")

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+consts.TestUser1ID+"/withdrawals", nil)
	req = withURLParam(req, "userID", consts.TestUser1ID)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Len(t, result, 2)
}

func TestListWithdrawals_ExcludesDeposits(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	deposit := &models.Transaction{
		UserID:      consts.TestUser1ID,
		Amount:      "100.00",
		TotalAmount: "100.00",
		Fee:         "0.00",
		Currency:    "USD",
		DepositType: consts.DepositTypeExternal,
		Type:        consts.TransactionTypeDeposit,
		Status:      consts.TransactionStatusPending,
	}
	require.NoError(t, store.CreateTransaction(deposit))

	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+consts.TestUser1ID+"/withdrawals", nil)
	req = withURLParam(req, "userID", consts.TestUser1ID)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Empty(t, result)
}

// ── TriggerWithdrawalEvent ──

func TestTriggerWithdrawalEvent_Completed(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")

	body, _ := json.Marshal(map[string]string{"event": consts.WebhookEventWithdrawalCompleted})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/"+txID+"/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", txID)
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	bal, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.InDelta(t, 9950.0, bal, 0.01)
	tx, err := store.GetTransaction(txID)
	require.NoError(t, err)
	assert.Equal(t, consts.TransactionStatusCompleted, tx.Status)
}

func TestTriggerWithdrawalEvent_Rejected(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")

	body, _ := json.Marshal(map[string]string{"event": consts.WebhookEventWithdrawalRejected})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/"+txID+"/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", txID)
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	bal, _ := store.GetBalance(consts.TestUser1ID, "USD")
	assert.Equal(t, 10000.0, bal)
	tx, err := store.GetTransaction(txID)
	require.NoError(t, err)
	assert.Equal(t, consts.TransactionStatusFailed, tx.Status)
}

func TestTriggerWithdrawalEvent_NotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	body, _ := json.Marshal(map[string]string{"event": consts.WebhookEventWithdrawalCompleted})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/nonexistent/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", "nonexistent")
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTriggerWithdrawalEvent_NotPending(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")
	require.NoError(t, store.UpdateTransactionStatus(txID, consts.TransactionStatusCompleted))

	body, _ := json.Marshal(map[string]string{"event": consts.WebhookEventWithdrawalCompleted})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/"+txID+"/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", txID)
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestTriggerWithdrawalEvent_UnsupportedEvent(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	txID := createPendingWithdrawal(t, store, consts.TestUser1ID, "USD", "50.00")

	body, _ := json.Marshal(map[string]string{"event": "unknown.event"})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/"+txID+"/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", txID)
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTriggerWithdrawalEvent_NotAWithdrawal(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	h := NewHandler(store, webhook.NewManager("", "s", nil, store, ""))

	deposit := &models.Transaction{
		UserID:      consts.TestUser1ID,
		Amount:      "100.00",
		TotalAmount: "100.00",
		Fee:         "0.00",
		Currency:    "USD",
		DepositType: consts.DepositTypeExternal,
		Type:        consts.TransactionTypeDeposit,
		Status:      consts.TransactionStatusPending,
	}
	require.NoError(t, store.CreateTransaction(deposit))

	body, _ := json.Marshal(map[string]string{"event": consts.WebhookEventWithdrawalCompleted})
	req := httptest.NewRequest(http.MethodPost, "/admin/withdrawals/"+deposit.ID+"/trigger-event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "txID", deposit.ID)
	rec := httptest.NewRecorder()
	h.TriggerWithdrawalEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
