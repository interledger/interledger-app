package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mockxago/internal/logger"
)

// ListTransactions handles GET /company/transactions?limit=10&page=1
// Returns a paginated list of transactions
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	limit := 10
	page := 1

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Calculate offset from page number
	offset := (page - 1) * limit

	// Get deposits from storage with pagination
	deposits, total, err := h.store.ListDeposits(r.Context(), limit, offset)
	if err != nil {
		logger.Errorf("Failed to list transactions: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve transactions")
		return
	}

	// Convert deposits to response format
	data := make([]map[string]interface{}, 0)
	if deposits != nil {
		for _, deposit := range deposits {
			data = append(data, map[string]interface{}{
				"transactionId": deposit.ID,
				"accountId":     deposit.AccountID,
				"amount":        deposit.Amount,
				"currencyCode":  deposit.Currency,
				"status":        deposit.Status,
				"code":          deposit.Code,
				"createdAt":     deposit.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				"settledAt":     deposit.SettledAt,
				"isDuplicate":   false,
				"isRequested":   false,
			})
		}
	}

	// Build response with transaction list
	response := map[string]interface{}{
		"data": data,
		"pagination": map[string]interface{}{
			"limit": limit,
			"page":  page,
			"total": total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	logger.Infof("Listed transactions: page=%d, limit=%d, total=%d, returned=%d", page, limit, total, len(data))
}

// GetTransaction handles GET /company/transactions/{id}
// Returns a transaction that was previously created (for backend verification)
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "id")

	if txID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "transaction ID is required")
		return
	}

	// Try to retrieve the deposit/transaction from storage
	deposit, err := h.store.GetDeposit(r.Context(), txID)
	if err != nil {
		logger.Warnf("Transaction not found: %s", txID)
		h.sendError(w, http.StatusNotFound, "not_found", "transaction not found")
		return
	}

	// Build response matching backend expectations
	response := map[string]interface{}{
		"transactionId": deposit.ID,
		"accountId":     deposit.AccountID,
		"amount":        deposit.Amount,
		"currencyCode":  deposit.Currency,
		"status":        deposit.Status,
		"code":          deposit.Code,
		"createdAt":     deposit.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"settledAt":     deposit.SettledAt,
		"isDuplicate":   false,
		"isRequested":   false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	logger.Infof("Transaction retrieved: %s", txID)
}
