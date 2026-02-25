package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mockxago/internal/jobs"
	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
)

func setupTestHandler(t *testing.T) *Handler {
	// Set environment variables for testing
	os.Setenv("XAGO_API_PUBLIC_KEY", "test-public-key")
	os.Setenv("XAGO_API_SECRET", "test-secret")
	os.Setenv("XAGO_MOCK_TEST_MODE", "true")

	store := storage.NewMemoryStorage()
	queue := jobs.NewQueue(store)
	return NewHandler(store, queue)
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
