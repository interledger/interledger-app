package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/consts"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStatementsHandler(t *testing.T) *Handler {
	t.Helper()
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	return NewHandler(store, wm)
}

func statementChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- GetAccountConfirmation ---

func TestGetAccountConfirmation_MissingWalletAddress(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/account-confirmation/", nil)
	w := httptest.NewRecorder()
	h.GetAccountConfirmation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAccountConfirmation_Success(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/account-confirmation/rWallet123", nil)
	req = statementChiParams(req, map[string]string{"walletAddress": "rWallet123"})
	w := httptest.NewRecorder()
	h.GetAccountConfirmation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

// --- GetAccountStatement ---

func TestGetAccountStatement_MissingWalletAddress(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/rWallet/2024/01", nil)
	w := httptest.NewRecorder()
	h.GetAccountStatement(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAccountStatement_MissingYear(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/rWallet123//01", nil)
	req = statementChiParams(req, map[string]string{"walletAddress": "rWallet123", "month": "01"})
	w := httptest.NewRecorder()
	h.GetAccountStatement(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAccountStatement_MissingMonth(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/rWallet123/2024/", nil)
	req = statementChiParams(req, map[string]string{"walletAddress": "rWallet123", "year": "2024"})
	w := httptest.NewRecorder()
	h.GetAccountStatement(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAccountStatement_Success(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/rWallet123/2024/01", nil)
	req = statementChiParams(req, map[string]string{
		"walletAddress": "rWallet123",
		"year":          "2024",
		"month":         "01",
	})
	w := httptest.NewRecorder()
	h.GetAccountStatement(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

// --- GetTransferConfirmation ---

func TestGetTransferConfirmation_MissingTransactionUUID(t *testing.T) {
	h := setupStatementsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/transfer-confirmation/", nil)
	w := httptest.NewRecorder()
	h.GetTransferConfirmation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTransferConfirmation_Success(t *testing.T) {
	h := setupStatementsHandler(t)

	err := h.store.CreateTransaction(&models.Transaction{ID: "tx-uuid-123", Type: consts.TransactionTypeDeposit})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/statement/v1/statements/transfer-confirmation/tx-uuid-123", nil)
	req = statementChiParams(req, map[string]string{"transactionUUID": "tx-uuid-123"})
	w := httptest.NewRecorder()
	h.GetTransferConfirmation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}
