package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
	"gitlab.com/fynbos/mock/mockpti/internal/utils"
)

// transactionStatus determines the final status of a new transaction based on
// the x-pti-scenario-id header. By default transactions settle immediately.
// Callers can trigger failure paths by including FAIL, REFUSE, ERROR, or
// CANCEL (case-insensitive) in the scenario id.
func transactionStatus(scenarioID string) string {
	upper := strings.ToUpper(scenarioID)
	switch {
	case strings.Contains(upper, "REFUSE") || strings.Contains(upper, "FAIL"):
		return "REFUSED"
	case strings.Contains(upper, "ERROR"):
		return "ERROR"
	case strings.Contains(upper, "CANCEL"):
		return "CANCELED"
	default:
		return "SETTLED"
	}
}

// buildTransaction creates a Transaction model from common fields.
func (h *Handler) buildTransaction(r *http.Request, txType, userID, currency string, amount float64) *models.Transaction {
	requestID := r.Header.Get("x-pti-request-id")
	if requestID == "" {
		requestID = utils.GenerateUUID()
	}
	scenarioID := r.Header.Get("x-pti-scenario-id")
	return &models.Transaction{
		RequestID:       requestID,
		Status:          transactionStatus(scenarioID),
		TransactionType: txType,
		Amount:          amount,
		Currency:        currency,
		Date:            time.Now().Format(time.RFC3339),
		UserID:          userID,
		ResourceType:    "TRANSACTION_STATUS",
		ClientID:        h.config.ClientID,
	}
}

// CreateDeposit handles POST /transactions/deposits.
func (h *Handler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	currency := req.SourceMethod.Currency
	if currency == "" {
		currency = "USD"
	}

	tx := h.buildTransaction(r, "DEPOSIT", req.Initiator.ID, currency, req.Amount)

	if err := h.store.SaveTransaction(r.Context(), tx); err != nil {
		logger.Errorf("failed to save deposit transaction: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create deposit")
		return
	}

	logger.Infof("Created deposit transaction requestId=%s status=%s", tx.RequestID, tx.Status)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   tx.RequestID,
		Link: "/transactions/" + tx.RequestID,
	})
}

// CreateWithdrawal handles POST /transactions/withdrawals.
func (h *Handler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	currency := req.DestinationMethod.Currency
	if currency == "" {
		currency = "USD"
	}

	tx := h.buildTransaction(r, "WITHDRAWAL", req.Initiator.ID, currency, req.Amount)

	if err := h.store.SaveTransaction(r.Context(), tx); err != nil {
		logger.Errorf("failed to save withdrawal transaction: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create withdrawal")
		return
	}

	logger.Infof("Created withdrawal transaction requestId=%s status=%s", tx.RequestID, tx.Status)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   tx.RequestID,
		Link: "/transactions/" + tx.RequestID,
	})
}

// CreateTransfer handles POST /transactions/transfers.
func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	tx := h.buildTransaction(r, "TRANSFER", req.Initiator.ID, "USD", req.Amount)

	if err := h.store.SaveTransaction(r.Context(), tx); err != nil {
		logger.Errorf("failed to save transfer transaction: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create transfer")
		return
	}

	logger.Infof("Created transfer transaction requestId=%s status=%s", tx.RequestID, tx.Status)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   tx.RequestID,
		Link: "/transactions/" + tx.RequestID,
	})
}

// GetTransaction handles GET /transactions/{requestId}.
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	if requestID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "requestId is required")
		return
	}

	tx, err := h.store.GetTransaction(r.Context(), requestID)
	if err != nil {
		if errors.Is(err, storage.ErrTransactionNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "transaction not found")
			return
		}
		logger.Errorf("failed to get transaction %s: %v", requestID, err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get transaction")
		return
	}

	h.sendJSON(w, http.StatusOK, tx)
}

// UpdateTransaction handles POST /transactions/{requestId}/updates.
func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	if requestID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "requestId is required")
		return
	}

	// Verify the transaction exists before accepting the update.
	if _, err := h.store.GetTransaction(r.Context(), requestID); err != nil {
		if errors.Is(err, storage.ErrTransactionNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "transaction not found")
			return
		}
		logger.Errorf("failed to get transaction for update %s: %v", requestID, err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get transaction")
		return
	}

	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	updateID := utils.GenerateUUID()
	update := &models.TransactionUpdate{
		ID:            updateID,
		RequestID:     requestID,
		TransactionID: req.TransactionID,
		Feedback:      req.Feedback,
		Date:          time.Now(),
		ProviderName:  req.ProviderName,
		Payload:       req.Payload,
	}

	if err := h.store.SaveTransactionUpdate(r.Context(), update); err != nil {
		logger.Errorf("failed to save transaction update: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to save update")
		return
	}

	logger.Infof("Saved transaction update requestId=%s updateId=%s", requestID, updateID)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   updateID,
		Link: "/transactions/" + requestID + "/updates/" + updateID,
	})
}
