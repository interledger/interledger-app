package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockxago/internal/jobs"
	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

// JobTypeProcessDeposit is the job type for async deposit processing.
const JobTypeProcessDeposit = "process_deposit"

// SimulateTestDeposit handles POST /v1/company/accounts/testdeposit
// This is a public endpoint (no auth required) for simulating external deposits
func (h *Handler) SimulateTestDeposit(w http.ResponseWriter, r *http.Request) {
	var req models.TestDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.AccountID) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "accountId is required")
		return
	}
	if req.Amount < 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be positive")
		return
	}
	if req.Amount == 0 {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "amount must be greater than 0")
		return
	}
	if strings.TrimSpace(req.CurrencyCode) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "currencyCode is required")
		return
	}

	// Verify account exists
	_, err := h.store.GetSubAccount(r.Context(), req.AccountID)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "account not found")
		return
	}

	// Generate transaction ID
	transactionID := generateID()

	// Generate deposit reference if not provided
	depositReference := req.DepositReference
	if depositReference == "" {
		depositReference = "dep_" + transactionID
	}

	// Create deposit record with pending status
	deposit := &models.Deposit{
		ID:               transactionID,
		AccountID:        req.AccountID,
		Amount:           req.Amount,
		Currency:         req.CurrencyCode,
		DepositReference: depositReference,
		Status:           "pending",
		Code:             100, // Code 100 = pending
		CreatedAt:        time.Now(),
	}

	if err := h.store.SaveDeposit(r.Context(), deposit); err != nil {
		logger.Errorf("Failed to save deposit: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to create deposit")
		return
	}

	// Return pending response immediately
	resp := models.TestDepositResponse{
		TransactionID: transactionID,
		Status:        "pending",
	}
	h.sendJSON(w, http.StatusOK, resp)

	logger.Infof("Test deposit created: %s for account %s: %f %s (ref: %s)",
		transactionID, req.AccountID, req.Amount, req.CurrencyCode, depositReference)

	// Enqueue deposit processing job (will be picked up by the worker)
	jobData := map[string]interface{}{
		"transaction_id":    transactionID,
		"account_id":        req.AccountID,
		"amount":            req.Amount,
		"currency":          req.CurrencyCode,
		"deposit_reference": depositReference,
	}
	readyAt := time.Now().Add(500 * time.Millisecond) // simulate processing delay
	if _, err := h.queue.Enqueue(JobTypeProcessDeposit, jobData, readyAt); err != nil {
		logger.Errorf("Failed to enqueue deposit job: %v", err)
	}
}

// NewProcessDepositHandler returns a JobHandler that processes deposit jobs.
// It credits the balance, updates deposit status, and sends a webhook.
func (h *Handler) NewProcessDepositHandler() jobs.JobHandler {
	return func(ctx context.Context, job *models.Job) error {
		transactionID, _ := job.Data["transaction_id"].(string)
		accountID, _ := job.Data["account_id"].(string)
		amount, _ := job.Data["amount"].(float64)
		currency, _ := job.Data["currency"].(string)
		depositReference, _ := job.Data["deposit_reference"].(string)

		logger.Infof("Processing deposit job: txn=%s account=%s amount=%f %s",
			transactionID, accountID, amount, currency)

		// Get sub-account to find wallet ID
		subAccount, err := h.store.GetSubAccount(ctx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get sub-account %s: %w", accountID, err)
		}

		// Credit the balance
		logger.Infof("Crediting balance: walletID=%s, currency=%s, amount=%f", subAccount.WalletID, currency, amount)
		if err := h.store.AddBalance(ctx, subAccount.WalletID, currency, amount); err != nil {
			return fmt.Errorf("failed to credit balance for deposit %s: %w", transactionID, err)
		}
		logger.Infof("Balance credited successfully for deposit %s", transactionID)

		// Update deposit status to completed
		if err := h.store.UpdateDepositStatus(ctx, transactionID, "completed"); err != nil {
			logger.Errorf("Failed to update deposit status: %v", err)
		}

		// Update deposit code to 104 (successful)
		deposit, err := h.store.GetDeposit(ctx, transactionID)
		if err == nil {
			deposit.Code = 104
			deposit.Status = "completed"
			settledAt := time.Now()
			deposit.SettledAt = &settledAt
			if err := h.store.SaveDeposit(ctx, deposit); err != nil {
				logger.Errorf("Failed to update deposit record: %v", err)
			}
		}

		logger.Infof("Deposit completed: %s for account %s: %f %s",
			transactionID, accountID, amount, currency)

		// Send webhook notification
		h.sendDepositCompletedWebhook(accountID, amount, currency, transactionID, depositReference)

		return nil
	}
}

// sendDepositCompletedWebhook sends a webhook notification when a deposit completes
func (h *Handler) sendDepositCompletedWebhook(accountID string, amount float64, currency, transactionID, depositReference string) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		logger.Warnf("WEBHOOK_URL not configured, skipping webhook for deposit %s", transactionID)
		return
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	logger.Infof("Attempting to send webhook for deposit %s to %s", transactionID, webhookURL)

	// Retrieve deposit for complete details
	ctx := context.Background()
	deposit, err := h.store.GetDeposit(ctx, transactionID)
	if err != nil {
		logger.Errorf("Failed to retrieve deposit for webhook: %v", err)
		return
	}

	// Build webhook payload
	payload := models.DepositWebhookPayload{
		AccountID:            accountID,
		Amount:               amount,
		CurrencyCode:         currency,
		TransactionID:        transactionID,
		TransactionReference: depositReference,
		Status:               "completed",
		Code:                 104, // Code 104 = successful deposit
		CreatedAt:            deposit.CreatedAt.Format(time.RFC3339),
	}

	if deposit.SettledAt != nil {
		payload.SettledAt = deposit.SettledAt.Format(time.RFC3339)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}

	// Send webhook with proper headers
	req, err := http.NewRequest("POST", webhookURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		logger.Errorf("Failed to create webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-gatehub-app-id", "xago-mock")

	timestamp := time.Now().Unix()
	req.Header.Set("x-gatehub-timestamp", strconv.FormatInt(timestamp, 10))

	// Generate HMAC-SHA256 signature
	if webhookSecret != "" {
		signature := generateWebhookSignature(webhookSecret, timestamp, "POST", webhookURL, bodyBytes)
		req.Header.Set("x-gatehub-signature", signature)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send deposit webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Errorf("Deposit webhook returned status %d for transaction %s", resp.StatusCode, transactionID)
		return
	}

	logger.Infof("Deposit webhook sent successfully for transaction %s (account: %s, amount: %f %s)",
		transactionID, accountID, amount, currency)
}

// ListCompanyDeposits handles GET /v1/company/deposits
// Returns a paginated list of all deposits
func (h *Handler) ListCompanyDeposits(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
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

	// Calculate offset
	offset := (page - 1) * limit

	// List deposits from storage
	deposits, total, err := h.store.ListDeposits(r.Context(), limit, offset)
	if err != nil {
		logger.Errorf("Failed to list deposits: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve deposits")
		return
	}

	// Convert to API response format
	items := make([]models.DepositItem, 0, len(deposits))
	for _, d := range deposits {
		item := models.DepositItem{
			TransactionID:        d.ID,
			AccountID:            d.AccountID,
			Amount:               d.Amount,
			CurrencyCode:         d.Currency,
			DepositReference:     d.DepositReference,
			TransactionReference: d.DepositReference, // Same as deposit reference
			Status:               d.Status,
			Code:                 d.Code,
			CreatedAt:            d.CreatedAt.Format(time.RFC3339),
		}
		if d.SettledAt != nil {
			item.SettledAt = d.SettledAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}

	// Calculate pagination
	numberOfPages := (total + limit - 1) / limit
	if numberOfPages < 1 {
		numberOfPages = 1
	}

	resp := models.ListDepositsResponse{
		Data: items,
		Pagination: models.DepositPagination{
			Limit:         limit,
			Page:          page,
			NumberOfPages: numberOfPages,
			Total:         total,
		},
	}

	h.sendJSON(w, http.StatusOK, resp)
	logger.Infof("Listed deposits: page=%d, limit=%d, total=%d", page, limit, total)
}
