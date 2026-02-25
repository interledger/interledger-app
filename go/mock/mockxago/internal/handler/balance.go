package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
)

// GetBalance handles GET /xago/v1/accounts/{accountId}/balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
		return
	}
	if _, err := uuid.Parse(accountID); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
		return
	}

	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load account")
		return
	}

	currencies := []string{"ZAR", "USD"}
	balances := make([]models.BalanceItem, 0, len(currencies))
	logger.Infof("GetBalance: accountID=%s, walletID=%s", accountID, subAcc.WalletID)
	for _, cur := range currencies {
		available, reserved, err := h.store.GetBalance(r.Context(), subAcc.WalletID, cur)
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load balance")
			return
		}
		logger.Infof("Balance for %s: available=%f, reserved=%f", cur, available, reserved)
		balances = append(balances, models.BalanceItem{
			CurrencyCode: cur,
			Available:    available,
			Reserved:     reserved,
			Total:        available + reserved,
		})
	}

	h.sendJSON(w, http.StatusOK, models.BalanceResponse{
		AccountID: accountID,
		Balances:  balances,
	})
}
