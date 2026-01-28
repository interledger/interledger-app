package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mockxago/internal/models"
)

func createSubAccountForBalance(t *testing.T, h *Handler, token string, walletID string) models.CreateSubAccountResponse {
	t.Helper()
	req := models.CreateSubAccountRequest{
		WalletID:                  walletID,
		FirstName:                 "Balance",
		LastName:                  "User",
		Email:                     walletID + "@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Balance St",
		ThirdPartyVerificationURL: "https://app.withpersona.com/dashboard/inquiries/inq_bal",
	}
	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, body)
	w := httptest.NewRecorder()
	h.CreateSubAccount(w, r)

	var resp models.CreateSubAccountResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func requestBalance(t *testing.T, h *Handler, token string, accountID string) (*httptest.ResponseRecorder, models.BalanceResponse) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/xago/v1/accounts/"+accountID+"/balance", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", accountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetBalance(w, r)

	var resp models.BalanceResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	return w, resp
}

func TestBalance_InitialZero(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	created := createSubAccountForBalance(t, h, token, "wallet_bal_test")

	w, resp := requestBalance(t, h, token, created.AccountID)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, created.AccountID, resp.AccountID)
	assert.Len(t, resp.Balances, 2)
	for _, b := range resp.Balances {
		assert.Equal(t, 0.0, b.Available)
		assert.Equal(t, 0.0, b.Reserved)
		assert.Equal(t, 0.0, b.Total)
	}
}

func TestBalance_WithAvailableAndReserved(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	created := createSubAccountForBalance(t, h, token, "wallet_balance_reserved")

	_ = h.store.SetBalance(context.Background(), "wallet_balance_reserved", "ZAR", 5000.0, 500.0)

	w, resp := requestBalance(t, h, token, created.AccountID)
	assert.Equal(t, http.StatusOK, w.Code)

	var zar models.BalanceItem
	for _, b := range resp.Balances {
		if b.CurrencyCode == "ZAR" {
			zar = b
		}
	}
	assert.Equal(t, 5000.0, zar.Available)
	assert.Equal(t, 500.0, zar.Reserved)
	assert.Equal(t, 5500.0, zar.Total)
}

func TestBalance_DepositAccumulates(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)
	created := createSubAccountForBalance(t, h, token, "wallet_deposit")

	_ = h.store.AddBalance(context.Background(), "wallet_deposit", "ZAR", 1000.0)
	_ = h.store.AddBalance(context.Background(), "wallet_deposit", "ZAR", 2000.0)
	_ = h.store.AddBalance(context.Background(), "wallet_deposit", "ZAR", 1500.0)

	_, resp := requestBalance(t, h, token, created.AccountID)
	var zar models.BalanceItem
	for _, b := range resp.Balances {
		if b.CurrencyCode == "ZAR" {
			zar = b
		}
	}
	assert.Equal(t, 4500.0, zar.Available)
	assert.Equal(t, 0.0, zar.Reserved)
	assert.Equal(t, 4500.0, zar.Total)
}

func TestBalance_TransferReducesAvailable(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)
	created := createSubAccountForBalance(t, h, token, "wallet_transfer")

	_ = h.store.SetBalance(context.Background(), "wallet_transfer", "ZAR", 5000.0, 0)
	_ = h.store.SubtractBalance(context.Background(), "wallet_transfer", "ZAR", 2000.0)

	_, resp := requestBalance(t, h, token, created.AccountID)
	var zar models.BalanceItem
	for _, b := range resp.Balances {
		if b.CurrencyCode == "ZAR" {
			zar = b
		}
	}
	assert.Equal(t, 3000.0, zar.Available)
	assert.Equal(t, 3000.0, zar.Total)
}

func TestBalance_MultiCurrencyIsolation(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)
	created := createSubAccountForBalance(t, h, token, "wallet_multi")

	_ = h.store.SetBalance(context.Background(), "wallet_multi", "ZAR", 5000.0, 0)
	_ = h.store.SetBalance(context.Background(), "wallet_multi", "USD", 1000.0, 0)

	_, resp := requestBalance(t, h, token, created.AccountID)

	var zar, usd models.BalanceItem
	for _, b := range resp.Balances {
		switch b.CurrencyCode {
		case "ZAR":
			zar = b
		case "USD":
			usd = b
		}
	}

	assert.Equal(t, 5000.0, zar.Available)
	assert.Equal(t, 1000.0, usd.Available)
}

func TestBalance_InvalidAccountID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/accounts/invalid_123/balance", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "invalid_123")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetBalance(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBalance_RequiresAuth(t *testing.T) {
	h := setupAuthHandler(t)

	// Need an account ID to hit the route
	token := issueToken(t, h)
	created := createSubAccountForBalance(t, h, token, "wallet_noauth")

	r := httptest.NewRequest(http.MethodGet, "/xago/v1/accounts/"+created.AccountID+"/balance", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", created.AccountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.AuthMiddleware(http.HandlerFunc(h.GetBalance)).ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
