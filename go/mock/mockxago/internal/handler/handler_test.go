package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
)

// testConfig returns a minimal *config.Config suitable for unit tests.
func testConfig() *config.Config {
	return &config.Config{
		PublicKey: "test-public-key",
		Secret:    "test-secret",
		TestMode:  true,
	}
}

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	store := storage.NewMemoryStorage()
	queue := jobs.NewQueue(store)
	return NewHandler(store, queue, testConfig())
}

func TestLogin_Success(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotEmpty(t, resp.TokenValue)
}

func TestLogin_InvalidPublicKey(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "wrong-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "unauthorized", resp.Error)
}

func TestLogin_InvalidSecret(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "wrong-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_MissingPublicKey(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestLogin_MissingSecret(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_InvalidJSON(t *testing.T) {
	h := setupTestHandler(t)

	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "invalid_request", resp.Error)
}

func TestLogin_TokenPersistence(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	var resp models.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	token := resp.TokenValue

	// Verify token exists in storage
	retrievedToken, err := h.store.GetAccessToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, token, retrievedToken.Token)
}

func TestLogin_TokenExpiration(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, httpReq)

	var resp models.LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	token := resp.TokenValue

	// Verify token is not expired initially
	retrievedToken, err := h.store.GetAccessToken(context.Background(), token)
	assert.NoError(t, err)
	assert.False(t, retrievedToken.IsExpired())
}

func TestLogin_MultipleLogins(t *testing.T) {
	h := setupTestHandler(t)

	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}

	body, _ := json.Marshal(req)

	// First login
	httpReq1 := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	h.Login(w1, httpReq1)

	var resp1 models.LoginResponse
	json.NewDecoder(w1.Body).Decode(&resp1)
	token1 := resp1.TokenValue

	// Second login
	httpReq2 := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	h.Login(w2, httpReq2)

	var resp2 models.LoginResponse
	json.NewDecoder(w2.Body).Decode(&resp2)
	token2 := resp2.TokenValue

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Both should exist in storage
	t1, _ := h.store.GetAccessToken(context.Background(), token1)
	t2, _ := h.store.GetAccessToken(context.Background(), token2)
	assert.NotNil(t, t1)
	assert.NotNil(t, t2)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	h := setupTestHandler(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	h.AuthMiddleware(nextHandler).ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "unauthorized", resp.Error)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	h := setupTestHandler(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpReq.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	h.AuthMiddleware(nextHandler).ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	h := setupTestHandler(t)

	// Create an expired token
	expiredToken := &models.AccessToken{
		ID:        "expired-id",
		Token:     "expired-token",
		ExpiresAt: time.Unix(1, 0), // Unix epoch time 1 second - definitely expired
	}
	h.store.SaveAccessToken(context.Background(), expiredToken)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpReq.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()

	h.AuthMiddleware(nextHandler).ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	h := setupTestHandler(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpReq.Header.Set("Authorization", "Bearer nonexistent-token")
	w := httptest.NewRecorder()

	h.AuthMiddleware(nextHandler).ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	h := setupTestHandler(t)

	// Login to get a valid token
	req := models.LoginRequest{
		PolicyID: "test-policy",
		Fields: []models.FieldData{
			{FieldName: "publicKey", FieldValue: "test-public-key"},
			{FieldName: "secret", FieldValue: "test-secret"},
		},
	}
	body, _ := json.Marshal(req)
	loginReq := httptest.NewRequest("POST", "/xago/v1/login", bytes.NewReader(body))
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	var loginResp models.LoginResponse
	json.NewDecoder(loginW.Body).Decode(&loginResp)

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	httpReq.Header.Set("Authorization", "Bearer "+loginResp.TokenValue)
	w := httptest.NewRecorder()

	h.AuthMiddleware(nextHandler).ServeHTTP(w, httpReq)

	assert.True(t, nextHandlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
