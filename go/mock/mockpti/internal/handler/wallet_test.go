package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func createTestUser(t *testing.T, h *Handler) string {
	t.Helper()
	user := &models.User{
		ID:   "user-wallet-test",
		Type: "PERSON",
		Name: &models.Name{First: "Alice", Last: "Smith"},
	}
	if err := h.store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user.ID
}

func TestCreateWallet_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	body := models.CreateWalletRequest{
		ID:       "wallet-1",
		Currency: "USD",
		Type:     "FIAT",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/wallets", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.Wallet
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.WalletID != "wallet-1" {
		t.Errorf("expected walletId wallet-1, got %s", resp.WalletID)
	}
	if resp.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", resp.Currency)
	}
	if resp.CreateDateTime == "" {
		t.Error("expected non-empty createDateTime")
	}
}

func TestCreateWallet_GeneratesID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	body := models.CreateWalletRequest{
		Currency: "EUR",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/wallets", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.Wallet
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.WalletID == "" {
		t.Error("expected non-empty wallet id")
	}
}

func TestCreateWallet_UserNotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateWalletRequest{Currency: "USD"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/nonexistent/wallets", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCreateWallet_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/wallets", bytes.NewReader([]byte("bad json")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateWallet_MissingClientID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateWalletRequest{Currency: "USD"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users/user-1/wallets", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// No x-pti-client-id
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestListWallets_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	// Create two wallets
	for _, id := range []string{"w1", "w2"} {
		body := models.CreateWalletRequest{ID: id, Currency: "USD"}
		payload, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/wallets", bytes.NewReader(payload))
		ptiHeaders(req)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("setup: expected 200, got %d", rr.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/wallets", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var wallets []*models.Wallet
	_ = json.NewDecoder(rr.Body).Decode(&wallets)

	if len(wallets) != 2 {
		t.Errorf("expected 2 wallets, got %d", len(wallets))
	}
}

func TestListWallets_EmptyForNewUser(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/wallets", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var wallets []*models.Wallet
	_ = json.NewDecoder(rr.Body).Decode(&wallets)

	if len(wallets) != 0 {
		t.Errorf("expected 0 wallets, got %d", len(wallets))
	}
}

func TestGetWallet_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	// Create wallet
	body := models.CreateWalletRequest{ID: "wallet-get", Currency: "USD"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/users/"+userID+"/wallets", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Get wallet
	req = httptest.NewRequest(http.MethodGet, "/users/"+userID+"/wallets/wallet-get", nil)
	ptiHeaders(req)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var wallet models.Wallet
	_ = json.NewDecoder(rr.Body).Decode(&wallet)

	if wallet.WalletID != "wallet-get" {
		t.Errorf("expected walletId wallet-get, got %s", wallet.WalletID)
	}
	if wallet.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", wallet.Currency)
	}
}

func TestGetWallet_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)
	userID := createTestUser(t, h)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/wallets/nonexistent", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
