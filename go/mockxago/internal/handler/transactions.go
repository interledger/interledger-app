package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
	"gitlab.com/fynbos/mockxago/internal/utils"
	"go.uber.org/zap"
)

// CreateTransfer handles POST /v1/transfers
// Creates a transfer from a Xago account to an external beneficiary
func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Amount <= 0 {
		if req.Amount == 0 {
			h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be greater than 0")
		} else {
			h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be positive")
		}
		return
	}

	currency := strings.TrimSpace(req.CurrencyCode)
	if currency == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}

	beneficiaryID := strings.TrimSpace(req.BeneficiaryID)
	if beneficiaryID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "beneficiaryId is required")
		return
	}

	// Get beneficiary to validate it exists and get wallet ID
	beneficiary, err := h.store.GetBeneficiary(r.Context(), beneficiaryID)
	if err != nil {
		logger.Warn("Beneficiary not found for transfer",
			zap.String("beneficiary_id", beneficiaryID),
			zap.Error(err))
		h.sendError(w, http.StatusBadRequest, "invalid_request", "beneficiary not found")
		return
	}

	walletID := beneficiary.WalletID

	// Check if we already have a transaction for this idempotency key
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey != "" {
		existingTx, err := h.store.GetTransactionByIdempotencyKey(r.Context(), idempotencyKey)
		if err == nil && existingTx != nil {
			logger.Infof("Idempotent transfer request: returning existing transaction %s", existingTx.ID)
			h.sendJSON(w, http.StatusOK, models.CreateTransferResponse{
				TransactionID: existingTx.ID,
			})
			return
		}
	}

	// Check balance
	available, _, err := h.store.GetBalance(r.Context(), walletID, currency)
	if err != nil {
		logger.Error("Failed to get balance for transfer",
			zap.String("wallet_id", walletID),
			zap.String("currency", currency),
			zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to check balance")
		return
	}

	if available < req.Amount {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "insufficient balance")
		return
	}

	// Deduct balance
	if err := h.store.SubtractBalance(r.Context(), walletID, currency, req.Amount); err != nil {
		if err == storage.ErrInsufficientBalance {
			h.sendError(w, http.StatusBadRequest, "invalid_request", "insufficient balance")
		} else {
			logger.Error("Failed to subtract balance for transfer",
				zap.String("wallet_id", walletID),
				zap.Float64("amount", req.Amount),
				zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to process transfer")
		}
		return
	}

	// Create transaction
	transaction := &models.Transaction{
		ID:            utils.GenerateUUID(),
		AccountID:     beneficiary.AccountID,
		WalletID:      walletID,
		BeneficiaryID: beneficiaryID,
		Amount:        req.Amount,
		Currency:      currency,
		Reference:     req.Reference,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	if err := h.store.SaveTransaction(r.Context(), transaction); err != nil {
		logger.Error("Failed to save transaction",
			zap.String("transaction_id", transaction.ID),
			zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to create transaction")
		return
	}

	// Save idempotency key if provided
	if idempotencyKey != "" {
		if err := h.store.SaveIdempotencyKey(r.Context(), idempotencyKey, transaction.ID); err != nil {
			logger.Warn("Failed to save idempotency key",
				zap.String("idempotency_key", idempotencyKey),
				zap.Error(err))
		}
	}

	logger.Info("Transfer created",
		zap.String("transaction_id", transaction.ID),
		zap.String("wallet_id", walletID),
		zap.Float64("amount", req.Amount),
		zap.String("currency", currency),
		zap.String("beneficiary_id", beneficiaryID))

	// Schedule auto-completion after 2-3 seconds
	// IMPORTANT: Use context.Background() instead of r.Context() because the HTTP request
	// context will be cancelled after the response is sent, before the goroutine completes.
	go func() {
		time.Sleep(2500 * time.Millisecond) // 2.5 seconds
		if err := h.store.UpdateTransactionStatus(context.Background(), transaction.ID, "completed"); err != nil {
			logger.Errorf("Failed to complete transfer: %v", err)
		} else {
			logger.Infof("Transfer auto-completed: %s", transaction.ID)
		}
	}()

	// Return transaction ID
	h.sendJSON(w, http.StatusOK, models.CreateTransferResponse{
		TransactionID: transaction.ID,
	})
}

// ListTransactions handles GET /company/transactions?limit=10&page=1
// Returns a paginated list of both deposits and transfers
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
	deposits, depositTotal, err := h.store.ListDeposits(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list deposits",
			zap.Int("limit", limit),
			zap.Int("page", page),
			zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve transactions")
		return
	}

	// Convert deposits to transaction items
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
			"total": depositTotal,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	logger.Infof("Listed transactions: page=%d, limit=%d, total=%d, returned=%d", page, limit, depositTotal, len(data))
}

// ListTransfers handles GET /v1/transfers?limit=10&page=1
// Returns a paginated list of transfers
func (h *Handler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	// Resolve account ID from the access token for account-scoped results

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

	offset := (page - 1) * limit

	tokenValue := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if tokenValue == "" {
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
		return
	}

	accountID, err := h.store.GetAccountIDByToken(r.Context(), tokenValue)
	if err != nil || accountID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "account not found")
		return
	}

	// Get transfers from storage
	transfers, total, err := h.store.ListTransactionsByAccount(r.Context(), accountID, limit, offset)
	if err != nil {
		logger.Errorf("Failed to list transfers: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve transfers")
		return
	}

	// Convert to response format
	transactionItems := make([]models.TransactionItem, 0)
	if transfers != nil {
		for _, tx := range transfers {
			transactionItems = append(transactionItems, models.TransactionItem{
				TransactionID: tx.ID,
				Status:        tx.Status,
				Amount:        tx.Amount,
				CurrencyCode:  tx.Currency,
				BeneficiaryID: tx.BeneficiaryID,
				Reference:     tx.Reference,
				CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				SettledAt:     formatSettledAt(tx.SettledAt),
			})
		}
	}

	response := models.ListTransactionsResponse{
		Data: transactionItems,
		Pagination: models.TransactionPagination{
			Limit:         limit,
			Page:          page,
			NumberOfPages: (total + limit - 1) / limit, // Ceil division
			Total:         total,
		},
	}

	h.sendJSON(w, http.StatusOK, response)
	logger.Infof("Listed transfers: page=%d, limit=%d, total=%d, returned=%d", page, limit, total, len(transactionItems))
}

// GetTransaction handles GET /company/transactions/{id} and GET /v1/transfers/{id}
// Returns a transaction that was previously created (for backend verification)
// Searches for both transfers and deposits
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "id")

	if txID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "transaction ID is required")
		return
	}

	// Try to get as a regular transfer first
	transfer, err := h.store.GetTransaction(r.Context(), txID)
	if err == nil && transfer != nil {
		h.sendJSON(w, http.StatusOK, models.GetTransactionResponse{
			TransactionID: transfer.ID,
			Status:        transfer.Status,
			Amount:        transfer.Amount,
			CurrencyCode:  transfer.Currency,
			BeneficiaryID: transfer.BeneficiaryID,
			Reference:     transfer.Reference,
			CreatedAt:     transfer.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SettledAt:     formatSettledAt(transfer.SettledAt),
		})
		logger.Infof("Transfer retrieved: %s", txID)
		return
	}

	// Try to retrieve as a deposit
	deposit, err := h.store.GetDeposit(r.Context(), txID)
	if err != nil {
		if h.testMode {
			now := time.Now()
			h.sendJSON(w, http.StatusOK, models.GetTransactionResponse{
				TransactionID: txID,
				Status:        "completed",
				Amount:        0,
				CurrencyCode:  "ZAR",
				CreatedAt:     now.Format("2006-01-02T15:04:05Z07:00"),
				SettledAt:     now.Format("2006-01-02T15:04:05Z07:00"),
			})
			logger.Infof("Stubbed transfer returned in test mode: %s", txID)
			return
		}

		logger.Warnf("Transaction not found: %s", txID)
		h.sendError(w, http.StatusNotFound, "not_found", "transaction not found")
		return
	}

	// Build response matching backend expectations for deposits
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
	logger.Infof("Deposit retrieved: %s", txID)
}

// formatSettledAt formats the settled at timestamp
func formatSettledAt(settledAt *time.Time) string {
	if settledAt == nil {
		return ""
	}
	return settledAt.Format("2006-01-02T15:04:05Z07:00")
}

// GetTransactionByQuery handles GET /v1/transactions?transactionId=X
// Queries transaction by ID from query parameter (for wallet backend compatibility)
func (h *Handler) GetTransactionByQuery(w http.ResponseWriter, r *http.Request) {
	txID := r.URL.Query().Get("transactionId")

	if txID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "transactionId query parameter is required")
		return
	}

	// Try to get as a regular transfer first
	transfer, err := h.store.GetTransaction(r.Context(), txID)
	if err == nil && transfer != nil {
		h.sendJSON(w, http.StatusOK, models.GetTransactionResponse{
			TransactionID: transfer.ID,
			Status:        transfer.Status,
			Amount:        transfer.Amount,
			CurrencyCode:  transfer.Currency,
			BeneficiaryID: transfer.BeneficiaryID,
			Reference:     transfer.Reference,
			CreatedAt:     transfer.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SettledAt:     formatSettledAt(transfer.SettledAt),
		})
		logger.Infof("Transfer retrieved via query: %s", txID)
		return
	}

	// Try to retrieve as a deposit
	deposit, err := h.store.GetDeposit(r.Context(), txID)
	if err != nil {
		if h.testMode {
			now := time.Now()
			h.sendJSON(w, http.StatusOK, models.GetTransactionResponse{
				TransactionID: txID,
				Status:        "completed",
				Amount:        0,
				CurrencyCode:  "ZAR",
				CreatedAt:     now.Format("2006-01-02T15:04:05Z07:00"),
				SettledAt:     now.Format("2006-01-02T15:04:05Z07:00"),
			})
			logger.Infof("Test mode: returning mock transaction for %s", txID)
			return
		}
		h.sendError(w, http.StatusNotFound, "not_found", "transaction not found")
		return
	}

	// Build response matching backend expectations for deposits
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
	logger.Infof("Deposit retrieved via query: %s", txID)
}
