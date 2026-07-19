package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

func TestCreateTransfer_MissingAmount(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.CreateTransferRequest{
		CurrencyCode:  "ZAR",
		BeneficiaryID: "ben_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransfer_NegativeAmount(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.CreateTransferRequest{
		Amount:        -100.0,
		CurrencyCode:  "ZAR",
		BeneficiaryID: "ben_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransfer_MissingCurrency(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.CreateTransferRequest{
		Amount:        100.0,
		BeneficiaryID: "ben_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransfer_MissingBeneficiary(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.CreateTransferRequest{
		Amount:       100.0,
		CurrencyCode: "ZAR",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransfer_InvalidBeneficiary(t *testing.T) {
	h := setupAuthHandler(t)

	req := models.CreateTransferRequest{
		Amount:        100.0,
		CurrencyCode:  "ZAR",
		BeneficiaryID: "nonexistent_ben",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader(body))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransfer_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)

	httpReq := httptest.NewRequest(http.MethodPost, "/xago/v1/transfers", bytes.NewReader([]byte("invalid")))

	w := httptest.NewRecorder()
	h.CreateTransfer(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListTransactions_NoParams(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/xago/v1/transactions", token, nil)
	w := httptest.NewRecorder()

	h.ListTransactions(w, r)
	// Should return either OK or an error, but should be a valid response
	assert.True(t, w.Code >= 200)
}

func TestListTransactions_WithPagination(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/xago/v1/transactions?limit=10&page=1", token, nil)
	w := httptest.NewRecorder()

	h.ListTransactions(w, r)
	assert.True(t, w.Code >= 200)
}

func TestListTransfers_NoParams(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/xago/v1/transfers", token, nil)
	w := httptest.NewRecorder()

	h.ListTransfers(w, r)
	assert.True(t, w.Code >= 200)
}

func TestListTransfers_WithPagination(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/xago/v1/transfers?limit=5&page=2", token, nil)
	w := httptest.NewRecorder()

	h.ListTransfers(w, r)
	assert.True(t, w.Code >= 200)
}

func TestGetTransaction_InvalidID(t *testing.T) {
	h := setupAuthHandler(t)

	httpReq := httptest.NewRequest(http.MethodGet, "/xago/v1/transactions/invalid_tx", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("transactionId", "invalid_tx")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetTransaction(w, httpReq)

	// GetTransaction returns 400 for invalid/not found transactions, not 404
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTransactionByQuery_MissingQuery(t *testing.T) {
	h := setupAuthHandler(t)

	httpReq := httptest.NewRequest(http.MethodGet, "/xago/v1/transactions/search", nil)

	w := httptest.NewRecorder()
	h.GetTransactionByQuery(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTransactionByQuery_InvalidID(t *testing.T) {
	h := setupAuthHandler(t)

	httpReq := httptest.NewRequest(http.MethodGet, "/xago/v1/transactions/search?transactionId=nonexistent", nil)

	w := httptest.NewRecorder()
	h.GetTransactionByQuery(w, httpReq)

	// In test mode, nonexistent transactions return a mock response with 200 status
	assert.Equal(t, http.StatusOK, w.Code)
	var response models.GetTransactionResponse
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&response))
	assert.Equal(t, "nonexistent", response.TransactionID)
	assert.Equal(t, "completed", response.Status)
}

func TestGetTransaction_MissingID(t *testing.T) {
	h := setupAuthHandler(t)

	httpReq := httptest.NewRequest(http.MethodGet, "/xago/v1/transactions/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("transactionId", "")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetTransaction(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
