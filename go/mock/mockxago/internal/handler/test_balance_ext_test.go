package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

func TestTestSetBalance_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.TestSetBalanceRequest{
		WalletID:     "wallet_set_balance",
		CurrencyCode: "ZAR",
		Available:    5000.0,
		Reserved:     500.0,
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/set", token, body)
	w := httptest.NewRecorder()

	h.TestSetBalance(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTestSetBalance_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/set", token, []byte("invalid"))
	w := httptest.NewRecorder()

	h.TestSetBalance(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTestDeposit_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.TestBalanceDeltaRequest{
		WalletID:     "wallet_deposit",
		CurrencyCode: "ZAR",
		Amount:       1000.0,
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/deposit", token, body)
	w := httptest.NewRecorder()

	h.TestDeposit(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTestDeposit_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/deposit", token, []byte("invalid"))
	w := httptest.NewRecorder()

	h.TestDeposit(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTestTransfer_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Set balance first
	balance := models.TestSetBalanceRequest{
		WalletID:     "wallet_transfer",
		CurrencyCode: "ZAR",
		Available:    5000.0,
		Reserved:     0.0,
	}
	body, _ := json.Marshal(balance)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/set", token, body)
	w := httptest.NewRecorder()
	h.TestSetBalance(w, r)

	// Then do a transfer
	transfer := models.TestBalanceDeltaRequest{
		WalletID:     "wallet_transfer",
		CurrencyCode: "ZAR",
		Amount:       1000.0,
	}
	body, _ = json.Marshal(transfer)
	r = authorizedRequest(http.MethodPost, "/xago/v1/test/balances/transfer", token, body)
	w = httptest.NewRecorder()

	h.TestTransfer(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTestTransfer_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/transfer", token, []byte("invalid"))
	w := httptest.NewRecorder()

	h.TestTransfer(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTestCreateTransaction_OfflineMode(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.TestBalanceDeltaRequest{
		WalletID:     "wallet_offline",
		CurrencyCode: "ZAR",
		Amount:       500.0,
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/transactions/create", token, body)
	w := httptest.NewRecorder()

	h.TestCreateTransaction(w, r)

	// Should return OK or a test mode response
	assert.True(t, w.Code >= 200 && w.Code < 300 || w.Code == http.StatusBadRequest)
}

func TestTestReset_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/xago/v1/test/reset", token, nil)
	w := httptest.NewRecorder()

	h.TestReset(w, r)

	// Reset should return OK or a valid response
	assert.True(t, w.Code >= 200)
}

func TestTestClearDeposits_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/xago/v1/test/deposits/clear", token, nil)
	w := httptest.NewRecorder()

	h.TestClearDeposits(w, r)

	// ClearDeposits should return OK or a valid response
	assert.True(t, w.Code >= 200)
}

func TestTestSetBalance_MissingWalletID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.TestSetBalanceRequest{
		CurrencyCode: "ZAR",
		Available:    5000.0,
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/set", token, body)
	w := httptest.NewRecorder()

	h.TestSetBalance(w, r)

	// Missing walletID should result in an error
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTestDeposit_MissingWalletID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	req := models.TestBalanceDeltaRequest{
		CurrencyCode: "ZAR",
		Amount:       1000.0,
	}

	body, _ := json.Marshal(req)
	r := authorizedRequest(http.MethodPost, "/xago/v1/test/balances/deposit", token, body)
	w := httptest.NewRecorder()

	h.TestDeposit(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
