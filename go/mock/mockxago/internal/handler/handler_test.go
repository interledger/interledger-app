package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
)

func setupTestHandler(t *testing.T) *Handler {
	// Set environment variables for testing
	os.Setenv("XAGO_API_PUBLIC_KEY", "test-public-key")
	os.Setenv("XAGO_API_SECRET", "test-secret")
	os.Setenv("XAGO_MOCK_TEST_MODE", "true")

	store := storage.NewMemoryStorage()
	return NewHandler(store)
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

// Sub-account tests

func TestCreateSubAccount_Success(t *testing.T) {
	h := setupTestHandler(t)

	req := models.CreateSubAccountRequest{
		WalletID:                  "wallet_123",
		FirstName:                 "John",
		LastName:                  "Doe",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://example.com",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/v1/company/accounts", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotEmpty(t, resp.AccountID)
	assert.NotEmpty(t, resp.DepositAddress)
	assert.NotZero(t, resp.DepositTag)
	assert.NotEmpty(t, resp.Beneficiaries)
	assert.Len(t, resp.Beneficiaries, 2)
}

func TestCreateSubAccount_MissingFirstName(t *testing.T) {
	h := setupTestHandler(t)

	req := models.CreateSubAccountRequest{
		// Missing FirstName
		LastName:                  "Doe",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://example.com",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/v1/company/accounts", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Message, "firstName is required")
}

func TestCreateSubAccount_MissingEmail(t *testing.T) {
	h := setupTestHandler(t)

	req := models.CreateSubAccountRequest{
		FirstName: "John",
		LastName:  "Doe",
		// Missing Email
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://example.com",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/v1/company/accounts", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Message, "email is required")
}

func TestUpdateSubAccount_Success(t *testing.T) {
	h := setupTestHandler(t)

	// First, create a sub-account
	createdResp := createTestSubAccount(t, h)

	// Now update it
	updateReq := models.UpdateSubAccountRequest{
		ThirdPartyVerificationURL: "https://example.com/updated",
		IDNumber:                  "9001011234568",
		PhysicalAddress:           "456 Oak Ave",
	}

	body, _ := json.Marshal(updateReq)
	httpReq := httptest.NewRequest("PUT", "/v1/company/accounts/"+createdResp.AccountID, bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")

	// Set chi URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", createdResp.AccountID)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.UpdateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, createdResp.AccountID, resp.AccountID)
	assert.Equal(t, "updated", resp.Status)
}

func TestUpdateSubAccount_InvalidID(t *testing.T) {
	h := setupTestHandler(t)

	updateReq := models.UpdateSubAccountRequest{
		ThirdPartyVerificationURL: "https://example.com/updated",
	}

	body, _ := json.Marshal(updateReq)
	httpReq := httptest.NewRequest("PUT", "/v1/company/accounts/invalid-id", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.UpdateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Message, "invalid account ID format")
}

func TestUpdateSubAccount_NotFound(t *testing.T) {
	h := setupTestHandler(t)

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	updateReq := models.UpdateSubAccountRequest{
		ThirdPartyVerificationURL: "https://example.com/updated",
	}

	body, _ := json.Marshal(updateReq)
	httpReq := httptest.NewRequest("PUT", "/v1/company/accounts/"+validUUID, bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")

	// Set chi URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", validUUID)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.UpdateSubAccount(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSubAccountByWallet_Success(t *testing.T) {
	h := setupTestHandler(t)

	// First, create a sub-account
	createdResp := createTestSubAccount(t, h)

	// Now retrieve it
	httpReq := httptest.NewRequest("GET", "/v1/company/accounts?walletId=wallet_123", nil)
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.GetSubAccountByWallet(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, createdResp.AccountID, resp.AccountID)
}

func TestGetSubAccountByWallet_MissingWalletId(t *testing.T) {
	h := setupTestHandler(t)

	httpReq := httptest.NewRequest("GET", "/v1/company/accounts", nil)
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.GetSubAccountByWallet(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Contains(t, resp.Message, "walletId is required")
}

func TestGetSubAccountByWallet_NotFound(t *testing.T) {
	h := setupTestHandler(t)

	httpReq := httptest.NewRequest("GET", "/v1/company/accounts?walletId=nonexistent", nil)
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.GetSubAccountByWallet(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Helper function to create a test sub-account
func createTestSubAccount(t *testing.T, h *Handler) models.CreateSubAccountResponse {
	req := models.CreateSubAccountRequest{
		WalletID:                  "wallet_123",
		FirstName:                 "John",
		LastName:                  "Doe",
		Email:                     "john@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Main St",
		ThirdPartyVerificationURL: "https://example.com",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/v1/company/accounts", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	h.CreateSubAccount(w, httpReq)

	var resp models.CreateSubAccountResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}
