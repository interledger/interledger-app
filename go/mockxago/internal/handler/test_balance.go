package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
)

func (h *Handler) ensureTestMode(w http.ResponseWriter) bool {
	if h.testMode {
		return true
	}
	h.sendError(w, http.StatusNotFound, "not_found", "endpoint not available")
	return false
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
