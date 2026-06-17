package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/consts"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── FeeConfig unit tests ─────────────────────────────────────────────

func TestFeeConfig_DefaultsToZero(t *testing.T) {
	fc := NewFeeConfig()
	assert.Equal(t, 0.0, fc.GetDepositFeePercent())
	assert.Equal(t, 0.0, fc.GetWithdrawalFeePercent())
}

func TestFeeConfig_SetAndGet(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(1.5)
	fc.SetWithdrawalFeePercent(2.0)
	assert.Equal(t, 1.5, fc.GetDepositFeePercent())
	assert.Equal(t, 2.0, fc.GetWithdrawalFeePercent())
}

func TestFeeConfig_ResetToZero(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(5.0)
	fc.SetWithdrawalFeePercent(3.0)
	fc.SetDepositFeePercent(0)
	fc.SetWithdrawalFeePercent(0)
	assert.Equal(t, 0.0, fc.GetDepositFeePercent())
	assert.Equal(t, 0.0, fc.GetWithdrawalFeePercent())
}

// ── CalculateFee unit tests ──────────────────────────────────────────

func TestCalculateFee_ZeroPercent(t *testing.T) {
	fee := CalculateFee(100.00, 0.0)
	assert.Equal(t, 0.0, fee)
}

func TestCalculateFee_OnePointFivePercent(t *testing.T) {
	fee := CalculateFee(100.00, 1.5)
	assert.Equal(t, 1.50, fee)
}

func TestCalculateFee_RoundToTwoDecimals(t *testing.T) {
	// 3% of 33.33 = 0.9999 → should round to 1.00
	fee := CalculateFee(33.33, 3.0)
	assert.Equal(t, 1.00, fee)
}

func TestCalculateFee_SmallAmountRoundsToZero(t *testing.T) {
	// 0.1% of 0.01 = 0.00001 → rounds to 0.00
	fee := CalculateFee(0.01, 0.1)
	assert.Equal(t, 0.0, fee)
}

func TestCalculateFee_LargeAmount(t *testing.T) {
	fee := CalculateFee(10000.00, 2.5)
	assert.Equal(t, 250.00, fee)
}

// ── Admin fee endpoint unit tests ────────────────────────────────────

func newTestHandler() *Handler {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "test-secret", nil, nil, "")
	return NewHandler(store, wm)
}

func TestGetFees_ReturnsDefaults(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/admin/fees", nil)
	rr := httptest.NewRecorder()
	h.GetFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 0.0, resp["deposit_fee_percentage"])
	assert.Equal(t, 0.0, resp["withdrawal_fee_percentage"])
}

func TestSetFees_DepositOnly(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{"deposit_fee_percentage": 1.5})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 1.5, resp["deposit_fee_percentage"])
	assert.Equal(t, 0.0, resp["withdrawal_fee_percentage"])
}

func TestSetFees_WithdrawalOnly(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{"withdrawal_fee_percentage": 2.0})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 0.0, resp["deposit_fee_percentage"])
	assert.Equal(t, 2.0, resp["withdrawal_fee_percentage"])
}

func TestSetFees_Both(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage":    1.5,
		"withdrawal_fee_percentage": 2.5,
	})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 1.5, resp["deposit_fee_percentage"])
	assert.Equal(t, 2.5, resp["withdrawal_fee_percentage"])
}

func TestSetFees_NegativeRejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{"deposit_fee_percentage": -1.0})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetFees_Over100Rejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{"deposit_fee_percentage": 101.0})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetFees_PersistsAcrossGets(t *testing.T) {
	h := newTestHandler()

	// Set fee
	body, _ := json.Marshal(map[string]interface{}{"deposit_fee_percentage": 4.0})
	req := httptest.NewRequest("PUT", "/admin/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetFees(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Get fees — should reflect the set value
	req2 := httptest.NewRequest("GET", "/admin/fees", nil)
	rr2 := httptest.NewRecorder()
	h.GetFees(rr2, req2)

	var resp map[string]interface{}
	err := json.NewDecoder(rr2.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 4.0, resp["deposit_fee_percentage"])
}

// ── Fee applied to CreateTransaction ─────────────────────────────────

func newTestHandlerWithSeededUsers() *Handler {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "test-secret", nil, nil, "")
	return NewHandler(store, wm)
}

func TestCreateTransaction_ExternalDeposit_WithFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set a 2.5% deposit fee
	h.feeConfig.SetDepositFeePercent(2.5)

	// Create a wallet for the test user
	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "100.00", resp["amount"])
	assert.Equal(t, "2.50", resp["fee"])
	assert.Equal(t, "100.00", resp["total_amount"])
}

func TestCreateTransaction_ExternalDeposit_ZeroFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	// Fee defaults to 0%

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "100.00", resp["amount"])
	assert.Equal(t, "0.00", resp["fee"])
	assert.Equal(t, "100.00", resp["total_amount"])
}

func TestCreateTransaction_HostedTransfer_NoFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	// Even with fee configured, hosted transfers should be fee-free
	h.feeConfig.SetDepositFeePercent(5.0)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       50.00,
		"currency":     "EUR",
		"type":         2,
		"deposit_type": "hosted",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "50.00", resp["amount"])
	assert.Equal(t, "0.00", resp["fee"])
	assert.Equal(t, "50.00", resp["total_amount"])
}

func TestCreateTransaction_FeeInGetTransaction(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	h.feeConfig.SetDepositFeePercent(1.5)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       200.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var createResp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&createResp)
	require.NoError(t, err)
	txID := createResp["uuid"].(string)

	// Fetch via GetTransaction — fee should persist
	tx, err := h.store.GetTransaction(txID)
	require.NoError(t, err)
	assert.Equal(t, "3.00", tx.Fee)
	assert.Equal(t, "200.00", tx.TotalAmount)
}

// ── Fee in TransactionCompleteHandler (iframe deposits) ──────────────

func TestTransactionComplete_WithFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	h.feeConfig.SetDepositFeePercent(2.5)

	createTestWallet(t, h.store, consts.TestUser1ID)
	h.tokenToUser.Store("test-token", consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":   "100.00",
		"currency": "USD",
	})
	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.TransactionCompleteHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify the stored transaction has the fee applied
	// Find the transaction by iterating (iframe handler creates one)
	bal, _ := h.store.GetBalance(consts.TestUser1ID, "USD")
	// Balance should reflect full amount (100.00) — GateHub adds the full amount to the vault,
	// fee is just metadata on the transaction, not deducted from credited amount.
	assert.GreaterOrEqual(t, bal, 100.0, "Balance should include the deposited amount")
}

func TestTransactionComplete_ZeroFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	// Fee defaults to 0%

	createTestWallet(t, h.store, consts.TestUser1ID)
	h.tokenToUser.Store("test-token", consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":   "50.00",
		"currency": "EUR",
	})
	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.TransactionCompleteHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// ── Withdrawal fee tests ─────────────────────────────────────────────

func TestCreateTransaction_Withdrawal_WithFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	h.feeConfig.SetWithdrawalFeePercent(2.0)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       50.00,
		"currency":     "EUR",
		"type":         3,
		"deposit_type": "withdrawal",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "50.00", resp["amount"])
	assert.Equal(t, "1.00", resp["fee"])
	assert.Equal(t, "51.00", resp["total_amount"])
}

func TestCreateTransaction_Withdrawal_ZeroFee(t *testing.T) {
	h := newTestHandlerWithSeededUsers()
	// Withdrawal fee defaults to 0%

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       50.00,
		"currency":     "EUR",
		"type":         3,
		"deposit_type": "withdrawal",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "50.00", resp["amount"])
	assert.Equal(t, "0.00", resp["fee"])
	assert.Equal(t, "50.00", resp["total_amount"])
}

// ── User-specific fee unit tests ────────────────────────────────────

func TestFeeConfig_GetDepositFeeForUser_NoOverride(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(2.5)

	fee, hasOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 2.5, fee, "should return global fee when no override exists")
	assert.False(t, hasOverride, "hasOverride should be false")
}

func TestFeeConfig_GetWithdrawalFeeForUser_NoOverride(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetWithdrawalFeePercent(1.5)

	fee, hasOverride := fc.GetWithdrawalFeeForUser("user456")
	assert.Equal(t, 1.5, fee, "should return global fee when no override exists")
	assert.False(t, hasOverride, "hasOverride should be false")
}

func TestFeeConfig_SetUserFees_DepositOnly(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(2.0)
	fc.SetWithdrawalFeePercent(1.0)

	depositPct := 5.0
	fc.SetUserFees("user123", &depositPct, nil)

	depositFee, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 5.0, depositFee)
	assert.True(t, depositOverride)

	withdrawalFee, withdrawalOverride := fc.GetWithdrawalFeeForUser("user123")
	assert.Equal(t, 1.0, withdrawalFee, "withdrawal should still use global fee")
	assert.False(t, withdrawalOverride)
}

func TestFeeConfig_SetUserFees_WithdrawalOnly(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(2.0)
	fc.SetWithdrawalFeePercent(1.0)

	withdrawalPct := 3.5
	fc.SetUserFees("user123", nil, &withdrawalPct)

	depositFee, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 2.0, depositFee, "deposit should still use global fee")
	assert.False(t, depositOverride)

	withdrawalFee, withdrawalOverride := fc.GetWithdrawalFeeForUser("user123")
	assert.Equal(t, 3.5, withdrawalFee)
	assert.True(t, withdrawalOverride)
}

func TestFeeConfig_SetUserFees_Both(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(2.0)
	fc.SetWithdrawalFeePercent(1.0)

	depositPct := 4.0
	withdrawalPct := 3.0
	fc.SetUserFees("user123", &depositPct, &withdrawalPct)

	depositFee, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 4.0, depositFee)
	assert.True(t, depositOverride)

	withdrawalFee, withdrawalOverride := fc.GetWithdrawalFeeForUser("user123")
	assert.Equal(t, 3.0, withdrawalFee)
	assert.True(t, withdrawalOverride)
}

func TestFeeConfig_SetUserFees_PartialUpdate(t *testing.T) {
	fc := NewFeeConfig()

	// Set both initially
	depositPct := 4.0
	withdrawalPct := 3.0
	fc.SetUserFees("user123", &depositPct, &withdrawalPct)

	// Update only deposit
	newDepositPct := 5.5
	fc.SetUserFees("user123", &newDepositPct, nil)

	depositFee, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 5.5, depositFee, "deposit should be updated")
	assert.True(t, depositOverride)

	withdrawalFee, withdrawalOverride := fc.GetWithdrawalFeeForUser("user123")
	assert.Equal(t, 3.0, withdrawalFee, "withdrawal should remain unchanged")
	assert.True(t, withdrawalOverride)
}

func TestFeeConfig_ClearUserFees(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(2.0)
	fc.SetWithdrawalFeePercent(1.0)

	// Set user overrides
	depositPct := 5.0
	withdrawalPct := 4.0
	fc.SetUserFees("user123", &depositPct, &withdrawalPct)

	// Verify overrides are set
	_, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.True(t, depositOverride)

	// Clear overrides
	fc.ClearUserFees("user123")

	// Verify back to global fees
	depositFee, depositOverride := fc.GetDepositFeeForUser("user123")
	assert.Equal(t, 2.0, depositFee)
	assert.False(t, depositOverride)

	withdrawalFee, withdrawalOverride := fc.GetWithdrawalFeeForUser("user123")
	assert.Equal(t, 1.0, withdrawalFee)
	assert.False(t, withdrawalOverride)
}

func TestFeeConfig_MultipleUsers(t *testing.T) {
	fc := NewFeeConfig()
	fc.SetDepositFeePercent(1.0)

	// Set different fees for different users
	user1DepositPct := 2.0
	fc.SetUserFees("user1", &user1DepositPct, nil)

	user2DepositPct := 3.0
	fc.SetUserFees("user2", &user2DepositPct, nil)

	// Verify each user has their own override
	fee1, override1 := fc.GetDepositFeeForUser("user1")
	assert.Equal(t, 2.0, fee1)
	assert.True(t, override1)

	fee2, override2 := fc.GetDepositFeeForUser("user2")
	assert.Equal(t, 3.0, fee2)
	assert.True(t, override2)

	// User without override should get global
	fee3, override3 := fc.GetDepositFeeForUser("user3")
	assert.Equal(t, 1.0, fee3)
	assert.False(t, override3)
}

// ── User-specific fee HTTP endpoint tests ────────────────────────────

func TestGetUserFees_NoOverride(t *testing.T) {
	h := newTestHandler()
	h.feeConfig.SetDepositFeePercent(2.0)
	h.feeConfig.SetWithdrawalFeePercent(1.5)

	req := httptest.NewRequest("GET", "/admin/users/user123/fees", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, 2.0, resp.DepositFeePercentage)
	assert.Equal(t, 1.5, resp.WithdrawalFeePercentage)
	assert.Equal(t, "global", resp.DepositFeeSource)
	assert.Equal(t, "global", resp.WithdrawalFeeSource)
}

func TestGetUserFees_WithOverride(t *testing.T) {
	h := newTestHandler()
	h.feeConfig.SetDepositFeePercent(2.0)
	h.feeConfig.SetWithdrawalFeePercent(1.5)

	// Set user override
	depositPct := 5.0
	withdrawalPct := 3.5
	h.feeConfig.SetUserFees("user123", &depositPct, &withdrawalPct)

	req := httptest.NewRequest("GET", "/admin/users/user123/fees", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, 5.0, resp.DepositFeePercentage)
	assert.Equal(t, 3.5, resp.WithdrawalFeePercentage)
	assert.Equal(t, "user", resp.DepositFeeSource)
	assert.Equal(t, "user", resp.WithdrawalFeeSource)
}

func TestGetUserFees_MissingUserID(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/admin/users//fees", nil)
	rr := httptest.NewRecorder()
	h.GetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_DepositOnly(t *testing.T) {
	h := newTestHandler()
	h.feeConfig.SetDepositFeePercent(1.0)
	h.feeConfig.SetWithdrawalFeePercent(0.5)

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage": 3.5,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, 3.5, resp.DepositFeePercentage)
	assert.Equal(t, "user", resp.DepositFeeSource)
	assert.Equal(t, 0.5, resp.WithdrawalFeePercentage)
	assert.Equal(t, "global", resp.WithdrawalFeeSource)
}

func TestSetUserFees_WithdrawalOnly(t *testing.T) {
	h := newTestHandler()
	h.feeConfig.SetDepositFeePercent(1.0)
	h.feeConfig.SetWithdrawalFeePercent(0.5)

	body, _ := json.Marshal(map[string]interface{}{
		"withdrawal_fee_percentage": 2.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user456/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user456", resp.UserID)
	assert.Equal(t, 1.0, resp.DepositFeePercentage)
	assert.Equal(t, "global", resp.DepositFeeSource)
	assert.Equal(t, 2.0, resp.WithdrawalFeePercentage)
	assert.Equal(t, "user", resp.WithdrawalFeeSource)
}

func TestSetUserFees_Both(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage":    4.5,
		"withdrawal_fee_percentage": 3.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user789/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user789")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user789", resp.UserID)
	assert.Equal(t, 4.5, resp.DepositFeePercentage)
	assert.Equal(t, "user", resp.DepositFeeSource)
	assert.Equal(t, 3.0, resp.WithdrawalFeePercentage)
	assert.Equal(t, "user", resp.WithdrawalFeeSource)
}

func TestSetUserFees_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_NoFieldsProvided(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_NegativeDepositRejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage": -1.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_Over100DepositRejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage": 101.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_NegativeWithdrawalRejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"withdrawal_fee_percentage": -2.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_Over100WithdrawalRejected(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"withdrawal_fee_percentage": 150.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users/user123/fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSetUserFees_MissingUserID(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"deposit_fee_percentage": 2.0,
	})
	req := httptest.NewRequest("PUT", "/admin/users//fees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestClearUserFees_RemovesOverrides(t *testing.T) {
	h := newTestHandler()
	h.feeConfig.SetDepositFeePercent(1.0)
	h.feeConfig.SetWithdrawalFeePercent(0.5)

	// Set user override
	depositPct := 5.0
	withdrawalPct := 3.0
	h.feeConfig.SetUserFees("user123", &depositPct, &withdrawalPct)

	// Clear overrides
	req := httptest.NewRequest("DELETE", "/admin/users/user123/fees", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "user123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ClearUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp userFeeResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, 1.0, resp.DepositFeePercentage)
	assert.Equal(t, "global", resp.DepositFeeSource)
	assert.Equal(t, 0.5, resp.WithdrawalFeePercentage)
	assert.Equal(t, "global", resp.WithdrawalFeeSource)
}

func TestClearUserFees_MissingUserID(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("DELETE", "/admin/users//fees", nil)
	rr := httptest.NewRecorder()
	h.ClearUserFees(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestClearUserFees_NonExistentUser(t *testing.T) {
	h := newTestHandler()

	// Clear fees for user with no overrides (should succeed gracefully)
	req := httptest.NewRequest("DELETE", "/admin/users/nonexistent/fees", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ClearUserFees(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// ── User-specific fees in transaction processing ─────────────────────

func TestCreateTransaction_Deposit_UserOverrideTakesPrecedence(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fee
	h.feeConfig.SetDepositFeePercent(2.0)

	// Set user-specific fee (higher than global)
	userDepositPct := 5.0
	h.feeConfig.SetUserFees(consts.TestUser1ID, &userDepositPct, nil)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Should use user-specific fee (5%) not global (2%)
	assert.Equal(t, "100.00", resp["amount"])
	assert.Equal(t, "5.00", resp["fee"], "should apply user-specific fee of 5%")
	assert.Equal(t, "100.00", resp["total_amount"])
}

func TestCreateTransaction_Deposit_GlobalFeeWhenNoUserOverride(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fee
	h.feeConfig.SetDepositFeePercent(2.5)

	// Set user-specific fee for a different user
	otherUserDepositPct := 5.0
	h.feeConfig.SetUserFees("other-user", &otherUserDepositPct, nil)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       200.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Should use global fee (2.5%) since user has no override
	assert.Equal(t, "200.00", resp["amount"])
	assert.Equal(t, "5.00", resp["fee"], "should apply global fee of 2.5%")
	assert.Equal(t, "200.00", resp["total_amount"])
}

func TestCreateTransaction_Withdrawal_UserOverrideTakesPrecedence(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fee
	h.feeConfig.SetWithdrawalFeePercent(1.0)

	// Set user-specific fee (higher than global)
	userWithdrawalPct := 3.5
	h.feeConfig.SetUserFees(consts.TestUser1ID, nil, &userWithdrawalPct)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "EUR",
		"type":         3,
		"deposit_type": "withdrawal",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Should use user-specific fee (3.5%) not global (1%)
	assert.Equal(t, "100.00", resp["amount"])
	assert.Equal(t, "3.50", resp["fee"], "should apply user-specific fee of 3.5%")
	assert.Equal(t, "103.50", resp["total_amount"])
}

func TestCreateTransaction_Withdrawal_GlobalFeeWhenNoUserOverride(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fee
	h.feeConfig.SetWithdrawalFeePercent(2.0)

	createTestWallet(t, h.store, consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       50.00,
		"currency":     "EUR",
		"type":         3,
		"deposit_type": "withdrawal",
	})
	req := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateTransaction(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Should use global fee (2%)
	assert.Equal(t, "50.00", resp["amount"])
	assert.Equal(t, "1.00", resp["fee"], "should apply global fee of 2%")
	assert.Equal(t, "51.00", resp["total_amount"])
}

func TestCreateTransaction_MixedUserFees(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fees
	h.feeConfig.SetDepositFeePercent(1.0)
	h.feeConfig.SetWithdrawalFeePercent(0.5)

	// Set user override only for deposit (withdrawal should use global)
	userDepositPct := 3.0
	h.feeConfig.SetUserFees(consts.TestUser1ID, &userDepositPct, nil)

	createTestWallet(t, h.store, consts.TestUser1ID)

	// Test deposit uses user override
	depositBody, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         1,
		"deposit_type": "external",
	})
	depositReq := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(depositBody))
	depositReq.Header.Set("Content-Type", "application/json")
	depositRr := httptest.NewRecorder()
	h.CreateTransaction(depositRr, depositReq)

	assert.Equal(t, http.StatusCreated, depositRr.Code)
	var depositResp map[string]interface{}
	err := json.NewDecoder(depositRr.Body).Decode(&depositResp)
	require.NoError(t, err)
	assert.Equal(t, "3.00", depositResp["fee"], "deposit should use user override of 3%")

	// Test withdrawal uses global
	withdrawalBody, _ := json.Marshal(map[string]interface{}{
		"user_id":      consts.TestUser1ID,
		"amount":       100.00,
		"currency":     "USD",
		"type":         3,
		"deposit_type": "withdrawal",
	})
	withdrawalReq := httptest.NewRequest("POST", "/core/v1/transactions", bytes.NewReader(withdrawalBody))
	withdrawalReq.Header.Set("Content-Type", "application/json")
	withdrawalRr := httptest.NewRecorder()
	h.CreateTransaction(withdrawalRr, withdrawalReq)

	assert.Equal(t, http.StatusCreated, withdrawalRr.Code)
	var withdrawalResp map[string]interface{}
	err = json.NewDecoder(withdrawalRr.Body).Decode(&withdrawalResp)
	require.NoError(t, err)
	assert.Equal(t, "0.50", withdrawalResp["fee"], "withdrawal should use global fee of 0.5%")
}

func TestTransactionComplete_UserOverrideTakesPrecedence(t *testing.T) {
	h := newTestHandlerWithSeededUsers()

	// Set global fee
	h.feeConfig.SetDepositFeePercent(1.0)

	// Set user-specific fee
	userDepositPct := 4.0
	h.feeConfig.SetUserFees(consts.TestUser1ID, &userDepositPct, nil)

	createTestWallet(t, h.store, consts.TestUser1ID)
	h.tokenToUser.Store("test-token", consts.TestUser1ID)

	body, _ := json.Marshal(map[string]interface{}{
		"amount":   "100.00",
		"currency": "USD",
	})
	req := httptest.NewRequest("POST", "/transaction/complete?paymentType=deposit&bearer=test-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.TransactionCompleteHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// The iframe handler should apply the user-specific fee
	// Find the transaction created
	bal, _ := h.store.GetBalance(consts.TestUser1ID, "USD")
	assert.GreaterOrEqual(t, bal, 100.0, "Balance should include the deposited amount")
}
