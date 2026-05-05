package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSeededHandler(t *testing.T) (*Handler, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	h := NewHandler(store, wm)
	return h, store
}

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// createWalletForUser creates a wallet via the store and returns its address.
func createWalletForUser(t *testing.T, store *storage.MemoryStorage, userID string) string {
	t.Helper()
	wallet := &models.Wallet{
		Address: "rTestAddr" + userID[:8],
		UserID:  userID,
		Name:    "Test Wallet",
		Type:    consts.WalletTypeStandard,
		Network: consts.NetworkXRPLedger,
	}
	require.NoError(t, store.CreateWallet(wallet))
	return wallet.Address
}

// ---- CreateWallet ----

func TestCreateWallet_Success(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"user_id": consts.TestUser1ID, "name": "My Wallet"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/core/v1/wallets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParams(req, map[string]string{})

	w := httptest.NewRecorder()
	h.CreateWallet(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var wallet models.Wallet
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &wallet))
	assert.NotEmpty(t, wallet.Address)
	assert.Equal(t, "My Wallet", wallet.Name)
}

func TestCreateWallet_MissingUserID(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"name": "My Wallet"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/wallets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParams(req, map[string]string{})

	w := httptest.NewRecorder()
	h.CreateWallet(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateWallet_UserNotFound(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"user_id": "nonexistent", "name": "Wallet"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/wallets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParams(req, map[string]string{})

	w := httptest.NewRecorder()
	h.CreateWallet(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateWallet_FromURLParam(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"name": "My Wallet"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/wallets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})

	w := httptest.NewRecorder()
	h.CreateWallet(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// ---- GetUserWallets ----

func TestGetUserWallets_AutoCreatesWallet(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})

	w := httptest.NewRecorder()
	h.GetUserWallets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.UserWalletsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Wallets, 1)
	assert.True(t, resp.Wallets[0].Primary)
}

func TestGetUserWallets_UserNotFound(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/users/nonexistent", nil)
	req = withChiParams(req, map[string]string{"userID": "nonexistent"})

	w := httptest.NewRecorder()
	h.GetUserWallets(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserWallets_ReturnsExisting(t *testing.T) {
	h, store := newSeededHandler(t)
	createWalletForUser(t, store, consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})

	w := httptest.NewRecorder()
	h.GetUserWallets(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.UserWalletsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Wallets, 1)
}

// ---- GetWallet ----

func TestGetWallet_Success(t *testing.T) {
	h, store := newSeededHandler(t)
	addr := createWalletForUser(t, store, consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/wallets/"+addr, nil)
	req = withChiParams(req, map[string]string{"walletID": addr})

	w := httptest.NewRecorder()
	h.GetWallet(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetWallet_NotFound(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/wallets/nonexistent", nil)
	req = withChiParams(req, map[string]string{"walletID": "nonexistent"})

	w := httptest.NewRecorder()
	h.GetWallet(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- GetWalletBalance ----

func TestGetWalletBalance_Success(t *testing.T) {
	h, store := newSeededHandler(t)
	addr := createWalletForUser(t, store, consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/wallets/"+addr+"/balances", nil)
	req = withChiParams(req, map[string]string{"walletID": addr})

	w := httptest.NewRecorder()
	h.GetWalletBalance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var balances []models.WalletBalanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &balances))
	assert.Len(t, balances, len(consts.SandboxCurrencies))
}

func TestGetWalletBalance_NotFound(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/wallets/notexist/balances", nil)
	req = withChiParams(req, map[string]string{"walletID": "notexist"})

	w := httptest.NewRecorder()
	h.GetWalletBalance(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- CreateTransaction ----

func TestCreateTransaction_SuccessCore(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{
		"user_id":  consts.TestUser1ID,
		"amount":   100.0,
		"currency": "USD",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var tx models.Transaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	assert.Equal(t, "100.00", tx.Amount)
	assert.Equal(t, "USD", tx.Currency)
}

func TestCreateTransaction_MissingUserIDCore(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"amount": 100.0, "currency": "USD"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransaction_InvalidAmountCore(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"user_id": consts.TestUser1ID, "amount": 0, "currency": "USD"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransaction_InferCurrencyFromVault(t *testing.T) {
	h, _ := newSeededHandler(t)

	vaultUUID := consts.SandboxVaultIDs["EUR"]
	body := map[string]interface{}{
		"user_id":    consts.TestUser1ID,
		"amount":     50.0,
		"vault_uuid": vaultUUID,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var tx models.Transaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	assert.Equal(t, "EUR", tx.Currency)
}

func TestCreateTransaction_NoCurrencyNoVault(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"user_id": consts.TestUser1ID, "amount": 10.0}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTransaction_UserNotFound(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"user_id": "no-such-user", "amount": 10.0, "currency": "USD"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateTransaction_FromManagedUserHeader(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{"amount": 25.0, "currency": "USD"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-gatehub-managed-user-uuid", consts.TestUser1ID)

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateTransaction_ResolveFromReceivingAddress(t *testing.T) {
	h, store := newSeededHandler(t)
	addr := createWalletForUser(t, store, consts.TestUser1ID)

	body := map[string]interface{}{"amount": 25.0, "currency": "USD", "receiving_address": addr}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateTransaction_HostedTransfer(t *testing.T) {
	h, _ := newSeededHandler(t)

	body := map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       30.0,
		"currency":     "USD",
		"type":         consts.TransactionTypeHosted,
		"deposit_type": consts.DepositTypeHosted,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/core/v1/transactions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateTransaction(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var tx models.Transaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	assert.Equal(t, consts.DepositTypeHosted, tx.DepositType)
	assert.Equal(t, "0.00", tx.Fee)
}

// ---- GetTransaction ----

func TestGetTransaction_SuccessCore(t *testing.T) {
	h, store := newSeededHandler(t)

	tx := &models.Transaction{
		UserID:   consts.TestUser1ID,
		Amount:   "50.00",
		Currency: "USD",
		Status:   1,
	}
	require.NoError(t, store.CreateTransaction(tx))

	req := httptest.NewRequest(http.MethodGet, "/core/v1/transactions/"+tx.ID, nil)
	req = withChiParams(req, map[string]string{"txID": tx.ID})

	w := httptest.NewRecorder()
	h.GetTransaction(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, tx.ID, resp["uuid"])
}

func TestGetTransaction_NotFoundCore(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/core/v1/transactions/nonexistent", nil)
	req = withChiParams(req, map[string]string{"txID": "nonexistent"})

	w := httptest.NewRecorder()
	h.GetTransaction(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- GetUserCurrencies ----

func TestGetUserCurrencies_WithBalance(t *testing.T) {
	h, store := newSeededHandler(t)
	h.tokenToUser.Store("test-bearer", consts.TestUser1ID)
	// User1 is seeded with 10000 USD
	_ = store.AddBalance(consts.TestUser1ID, "EUR", 500)

	req := httptest.NewRequest(http.MethodGet, "/api/user-currencies?bearer=test-bearer", nil)
	w := httptest.NewRecorder()
	h.GetUserCurrencies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.CurrenciesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp.Currencies, "USD")
	assert.Contains(t, resp.Currencies, "EUR")
}

func TestGetUserCurrencies_UnknownBearer(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user-currencies?bearer=unknown", nil)
	w := httptest.NewRecorder()
	h.GetUserCurrencies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.CurrenciesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, len(consts.SandboxCurrencies), len(resp.Currencies))
}

func TestGetUserCurrencies_MissingBearer(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/user-currencies", nil)
	w := httptest.NewRecorder()
	h.GetUserCurrencies(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserCurrencies_AuthorizationHeader(t *testing.T) {
	h, _ := newSeededHandler(t)
	h.tokenToUser.Store("auth-token", consts.TestUser1ID)

	req := httptest.NewRequest(http.MethodGet, "/api/user-currencies", nil)
	req.Header.Set("Authorization", "Bearer auth-token")
	w := httptest.NewRecorder()
	h.GetUserCurrencies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Rates ----

func TestGetCurrentRates_Core(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/rates/v1/rates?counter=USD", nil)
	w := httptest.NewRecorder()
	h.GetCurrentRates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "USD", resp["counter"])
	for currency := range consts.SandboxRates {
		_, exists := resp[currency]
		assert.True(t, exists, "missing rate for %s", currency)
	}
}

func TestGetCurrentRates_DefaultCounter(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/rates/v1/rates", nil)
	w := httptest.NewRecorder()
	h.GetCurrentRates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "USD", resp["counter"])
}

func TestGetVaults_Core(t *testing.T) {
	h, _ := newSeededHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/rates/v1/vaults", nil)
	w := httptest.NewRecorder()
	h.GetVaults(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.GetVaultsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Vaults, len(consts.SandboxVaultIDs))
}
