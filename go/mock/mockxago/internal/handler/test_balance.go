package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
	"gitlab.com/fynbos/mock/mockxago/internal/utils"
)

func (h *Handler) ensureTestMode(w http.ResponseWriter) bool {
	if h.testMode {
		return true
	}
	h.sendError(w, http.StatusNotFound, "not_found", "endpoint not available")
	return false
}

// TestCreateTransaction handles POST /xago/v1/test/transactions (test-mode only)
// Creates a transaction that the backend can verify via GET /company/transactions/{id}
func (h *Handler) TestCreateTransaction(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	var req struct {
		TransactionID string  `json:"transactionId"`
		AccountID     string  `json:"accountId"`
		Amount        float64 `json:"amount"`
		CurrencyCode  string  `json:"currencyCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.TransactionID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "transactionId is required")
		return
	}
	if req.AccountID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "accountId is required")
		return
	}
	if req.Amount <= 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be greater than zero")
		return
	}
	if req.CurrencyCode == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}

	// Create and store the deposit for backend to verify
	deposit := &models.Deposit{
		ID:        req.TransactionID,
		AccountID: req.AccountID,
		Amount:    req.Amount,
		Currency:  req.CurrencyCode,
		Status:    "settled",
		Code:      104, // Code 104 = successful deposit
		CreatedAt: time.Now(),
	}

	if err := h.store.SaveDeposit(r.Context(), deposit); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create transaction")
		return
	}

	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})
	logger.Infof("Test transaction created: %s for account %s: %f %s", req.TransactionID, req.AccountID, req.Amount, req.CurrencyCode)
}

func (h *Handler) resolveWalletID(r *http.Request, accountID string, walletID string) (string, error) {
	if strings.TrimSpace(walletID) != "" {
		return walletID, nil
	}
	if strings.TrimSpace(accountID) == "" {
		return "", storage.ErrSubAccountNotFound
	}
	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		return "", err
	}
	return subAcc.WalletID, nil
}

// TestSetBalance handles POST /xago/v1/test/balances/set (test-mode only).
func (h *Handler) TestSetBalance(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	var req models.TestSetBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	currency := strings.TrimSpace(req.CurrencyCode)
	if currency == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}

	walletID, err := h.resolveWalletID(r, req.AccountID, req.WalletID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
		return
	}

	if err := h.store.SetBalance(r.Context(), walletID, currency, req.Available, req.Reserved); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to set balance")
		return
	}

	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})
}

// TestDeposit handles POST /xago/v1/test/balances/deposit (test-mode only).
func (h *Handler) TestDeposit(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	var req models.TestBalanceDeltaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	currency := strings.TrimSpace(req.CurrencyCode)
	if currency == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}
	if req.Amount <= 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be greater than zero")
		return
	}

	walletID, err := h.resolveWalletID(r, req.AccountID, req.WalletID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
		return
	}

	if err := h.store.AddBalance(r.Context(), walletID, currency, req.Amount); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to apply deposit")
		return
	}

	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})

	// Send deposit webhook to backend (async) with optional transaction ID
	transactionID := req.TransactionID
	if transactionID == "" {
		transactionID = utils.GenerateUUID()
	}
	go h.sendDepositWebhook(walletID, currency, req.Amount, transactionID, req.AccountID)
}

// sendDepositWebhook sends a deposit webhook to the backend
func (h *Handler) sendDepositWebhook(walletID string, currency string, amount float64, transactionID string, accountID string) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	if webhookURL == "" {
		logger.Warnf("WEBHOOK_URL not configured, skipping webhook for wallet %s", walletID)
		return
	}

	// Use provided transaction and account IDs
	// If accountID is empty, lookup the sub-account for this wallet
	if accountID == "" {
		subAcc, err := h.store.GetSubAccountByWalletID(context.Background(), walletID)
		if err == nil && subAcc != nil {
			accountID = subAcc.AccountID
		} else {
			logger.Warnf("Could not find sub-account for wallet %s, cannot send webhook", walletID)
			return
		}
	}

	// Build webhook payload matching backend expectations
	// code: 104 = successful deposit, 100 = pending, etc.
	payload := map[string]interface{}{
		"accountId":        accountID,
		"amount":           amount,
		"currencyCode":     currency,
		"transactionId":    transactionID,
		"code":             104, // Code 104 indicates successful deposit
		"status":           "settled",
		"createdAt":        time.Now().Format(time.RFC3339),
		"settledAt":        time.Now().Format(time.RFC3339),
		"isDuplicate":      false,
		"isRequested":      false,
		"isRequestMatched": false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal deposit webhook: %v", err)
		return
	}

	// Send webhook with HMAC signature if secret is provided
	webhookReq, err := http.NewRequest("POST", webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		logger.Errorf("Failed to create webhook request: %v", err)
		return
	}

	webhookReq.Header.Set("Content-Type", "application/json")

	if webhookSecret != "" {
		h := hmac.New(sha256.New, []byte(webhookSecret))
		h.Write(bodyBytes)
		signature := hex.EncodeToString(h.Sum(nil))
		webhookReq.Header.Set("X-Signature", signature)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(webhookReq)
	if err != nil {
		logger.Errorf("Failed to send deposit webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Errorf("Deposit webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	logger.Infof("Deposit webhook sent successfully for wallet %s: %s %f (account: %s, tx: %s)",
		walletID, currency, amount, accountID, transactionID)
}

// TestTransfer handles POST /xago/v1/test/balances/transfer (test-mode only).
func (h *Handler) TestTransfer(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	var req models.TestBalanceDeltaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	currency := strings.TrimSpace(req.CurrencyCode)
	if currency == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}
	if req.Amount <= 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be greater than zero")
		return
	}

	walletID, err := h.resolveWalletID(r, req.AccountID, req.WalletID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
		return
	}

	if err := h.store.SubtractBalance(r.Context(), walletID, currency, req.Amount); err != nil {
		if err == storage.ErrInsufficientBalance {
			h.sendError(w, http.StatusBadRequest, "invalid_request", "insufficient balance")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to apply transfer")
		return
	}

	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})
}

// TestClearDeposits handles POST /v1/test/deposits/clear (test-mode only)
// Resets deposits, balances, and jobs to clear state between test scenarios,
// but preserves auth tokens and accounts so setup steps remain valid.
func (h *Handler) TestClearDeposits(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	if err := h.store.ClearDeposits(r.Context()); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to clear deposits")
		return
	}

	if err := h.store.ClearTransactions(r.Context()); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to clear transactions")
		return
	}

	if err := h.store.ClearBalances(r.Context()); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to clear balances")
		return
	}

	// Also clear jobs as they are related to deposits
	if err := h.store.ClearJobs(r.Context()); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to clear jobs")
		return
	}

	logger.Infof("Cleared deposits, balances and jobs (test reset)")
	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})
}

// TestReset handles POST /v1/test/reset (test-mode only)
// completely wipes all data from storage.
func (h *Handler) TestReset(w http.ResponseWriter, r *http.Request) {
	if !h.ensureTestMode(w) {
		return
	}

	if err := h.store.Reset(r.Context()); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to reset storage")
		return
	}

	logger.Infof("Reset ALL storage data (full wipe)")
	h.sendJSON(w, http.StatusOK, models.TestBalanceResponse{Status: "ok"})
}
