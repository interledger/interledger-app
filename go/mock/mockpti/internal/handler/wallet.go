package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/utils"
)

// CreateWallet handles POST /users/{id}/wallets.
func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id is required")
		return
	}

	// Verify user exists
	_, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	var req models.CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	walletID := req.ID
	if walletID == "" {
		walletID = utils.GenerateUUID()
	}

	wallet := &models.Wallet{
		WalletID:       walletID,
		Currency:       req.Currency,
		Reference:      req.Reference,
		CreateDateTime: time.Now().Format(time.RFC3339),
		Balance:        0,
		UserID:         userID,
	}

	if err := h.store.SaveWallet(r.Context(), wallet); err != nil {
		logger.Errorf("failed to save wallet: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create wallet")
		return
	}

	logger.Infof("Created wallet %s for user %s", walletID, userID)

	h.sendJSON(w, http.StatusOK, wallet)
}

// ListWallets handles GET /users/{id}/wallets.
func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id is required")
		return
	}

	wallets, err := h.store.ListWallets(r.Context(), userID)
	if err != nil {
		logger.Errorf("failed to list wallets: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to list wallets")
		return
	}

	h.sendJSON(w, http.StatusOK, wallets)
}

// GetWallet handles GET /users/{id}/wallets/{walletId}.
func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	walletID := chi.URLParam(r, "walletId")

	if userID == "" || walletID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id and wallet id are required")
		return
	}

	wallet, err := h.store.GetWallet(r.Context(), userID, walletID)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "wallet not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get wallet")
		return
	}

	h.sendJSON(w, http.StatusOK, wallet)
}
