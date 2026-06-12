package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/config"
	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	config         *config.Config
	store          storage.Storage
	webhookManager *webhook.Manager
	httpClient     *http.Client
	tokenToUser    sync.Map // Maps bearer tokens to user UUIDs
	feeConfig      *FeeConfig
}

// TransactionRequest represents a transaction request from the iframe
type TransactionRequest struct {
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	TriggerSCA bool   `json:"trigger_sca,omitempty"`
	TOTPCode   string `json:"totp_code,omitempty"`
}

// NewHandler creates a new handler using configuration loaded from the environment.
// For tests or custom setups, use NewHandlerWithConfig to inject a pre-built config.
func NewHandler(store storage.Storage, webhookManager *webhook.Manager) *Handler {
	return NewHandlerWithConfig(config.Load(), store, webhookManager)
}

// NewHandlerWithConfig creates a new handler with the provided configuration.
func NewHandlerWithConfig(cfg *config.Config, store storage.Storage, webhookManager *webhook.Manager) *Handler {
	if cfg == nil {
		cfg = config.Load()
	}
	logger.Info("initializing http handlers")
	return &Handler{
		config:         cfg,
		store:          store,
		webhookManager: webhookManager,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		feeConfig: NewFeeConfig(),
	}
}

// RequestLogger middleware logs all incoming requests with full details
func (h *Handler) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and re-buffer the body so downstream handlers can still read it
		var bodyStr string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				bodyStr = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Collect all headers into a map for structured logging
		headers := make(map[string]string, len(r.Header))
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ", ")
		}

		logger.Debug("request incoming",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Any("headers", headers),
			zap.String("body", bodyStr),
		)

		// Wrap ResponseWriter to capture status code
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.Debug("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// HealthCheck handles the health check endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Debug("health check requested")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"gitlab.com/fynbos/mock/mockgatehub"}`))
}

// RootHandler serves the main iframe page for deposit/onboarding
func (h *Handler) RootHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("root handler requested", zap.String("url", r.URL.String()), zap.String("host", r.Host), zap.String("referer", r.Header.Get("Referer")))

	paymentType := r.URL.Query().Get("paymentType")
	bearer := r.URL.Query().Get("bearer")

	if bearer == "" {
		logger.Error("missing bearer token in root request", zap.String("url", r.URL.String()))
		http.Error(w, "Missing bearer token", http.StatusBadRequest)
		return
	}

	logger.Info("serving iframe", zap.String("payment_type", paymentType), zap.String("bearer_prefix", tokenPrefix(bearer)))

	// If no paymentType is provided, treat this as onboarding and serve the KYC iframe
	if paymentType == "" || paymentType == "onboarding" {
		// Try to extract user UUID from bearer token mapping
		userUUID := h.extractUserFromBearer(bearer)

		// If not found in mapping, try to get from query params (user_id might be passed from frontend)
		if userUUID == "" {
			userUUID = r.URL.Query().Get("user_id")
		}

		if userUUID == "" {
			logger.Warn("could not extract user from bearer token or query params, will rely on form submission")
		}

		// Load KYC iframe template
		kycTemplatePath := filepath.Join("web", "kyc-iframe.html")
		kycTmpl, err := template.ParseFiles(kycTemplatePath)
		if err != nil {
			logger.Error("failed to parse kyc iframe template", zap.Error(err))
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		// Prepare data for KYC template - pass bearer token to be used on submission
		kycData := map[string]string{
			"Token":  bearer,
			"UserID": userUUID,
		}

		// Set headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Render KYC iframe
		if err := kycTmpl.Execute(w, kycData); err != nil {
			logger.Error("failed to execute kyc iframe template", zap.Error(err))
			http.Error(w, "Template execution error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Otherwise, serve the payment iframe (deposit/withdrawal/exchange)
	bearerShort := tokenPrefix(bearer)

	// Select the appropriate template based on payment type
	var templateFile string
	switch paymentType {
	case "deposit":
		templateFile = "deposit.html"
	case "withdrawal":
		templateFile = "withdrawal.html"
	default:
		templateFile = "index.html"
	}

	// Load template from web folder
	templatePath := filepath.Join("web", templateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Error("failed to parse template", zap.Error(err))
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Serialize consts for injection into the template as safe JavaScript
	vaultJSON, err := json.Marshal(consts.VaultUUIDToCurrency)
	if err != nil {
		logger.Error("failed to marshal vault UUID mapping", zap.Error(err))
		http.Error(w, "Template data preparation error", http.StatusInternalServerError)
		return
	}
	currenciesJSON, err := json.Marshal(consts.SandboxCurrencies)
	if err != nil {
		logger.Error("failed to marshal currencies", zap.Error(err))
		http.Error(w, "Template data preparation error", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := map[string]interface{}{
		"PaymentType":         paymentType,
		"Bearer":              bearer,
		"BearerShort":         bearerShort,
		"VaultUUIDToCurrency": template.JS(string(vaultJSON)),
		"AvailableCurrencies": template.JS(string(currenciesJSON)),
	}

	// Set headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Render template
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("failed to execute template", zap.Error(err))
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// TransactionCompleteHandler handles transaction completion callbacks from the iframe
func (h *Handler) TransactionCompleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	logger.Info("transaction complete handler requested")

	paymentType := r.URL.Query().Get("paymentType")

	// Try to get bearer from query parameter first (consistent with GetUserCurrencies)
	bearer := r.URL.Query().Get("bearer")

	// Fall back to Authorization header if not in query parameter
	if bearer == "" {
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			bearer = extractBearerFromAuthHeader(authHeader)
		}
	}

	if bearer == "" {
		logger.Error("missing bearer token in transaction completion")
		h.sendErrorWithCORS(w, http.StatusBadRequest, "Missing bearer token")
		return
	}

	logger.Info("transaction completed", zap.String("payment_type", paymentType))

	// Parse and validate request body for transaction details
	var txReq TransactionRequest
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &txReq); err != nil {
				logger.Warn("failed to parse request body", zap.Error(err))
			}
		}
	}

	// Validate and error out if required fields are missing
	if err := h.validateTransactionRequest(&txReq); err != nil {
		logger.Error("invalid transaction request", zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("parsed transaction details", zap.String("amount", txReq.Amount), zap.String("currency", txReq.Currency))

	// Optional SCA TOTP verification (same flow as KYC 2FA)
	if txReq.TriggerSCA {
		userUUID := h.extractUserFromBearer(bearer)
		if userUUID == "" {
			logger.Warn("SCA requested but could not extract user from bearer")
			h.sendErrorWithCORS(w, http.StatusBadRequest, "Invalid bearer token for SCA verification")
			return
		}

		if err := h.verify2FA(userUUID, txReq.TOTPCode); err != nil {
			h.sendErrorWithCORS(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("SCA TOTP verification succeeded",
			zap.String("user_id", userUUID),
			zap.String("payment_type", paymentType),
		)
	}

	// For deposit type, create a transaction and send a webhook to wallet-backend
	if paymentType == "deposit" {
		h.processDeposit(w, bearer, &txReq)
		return
	}

	// For withdrawal type, create a withdrawal transaction and send a webhook
	if paymentType == "withdraw" || paymentType == "withdrawal" {
		h.processWithdrawal(w, bearer, &txReq)
		return
	}

	// Return success response
	h.sendJSONWithCORS(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Transaction completed",
	})
}

// validateTransactionRequest checks that amount and currency are provided and valid
func (h *Handler) validateTransactionRequest(txReq *TransactionRequest) error {
	if txReq.Amount == "" {
		return fmt.Errorf("amount is required")
	}

	if txReq.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// Validate amount is numeric
	if _, err := strconv.ParseFloat(txReq.Amount, 64); err != nil {
		return fmt.Errorf("amount must be a valid number")
	}

	// Validate currency is supported
	if !isSupportedCurrency(txReq.Currency) {
		return fmt.Errorf("unsupported currency: %s", txReq.Currency)
	}

	return nil
}

// isSupportedCurrency checks if a currency is in the list of supported currencies
func isSupportedCurrency(currency string) bool {
	for _, c := range consts.SandboxCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// processDeposit handles the deposit transaction logic
func (h *Handler) processDeposit(w http.ResponseWriter, bearer string, txReq *TransactionRequest) {
	// Decode bearer to get user UUID
	userUUID := h.extractUserFromBearer(bearer)

	if userUUID == "" {
		logger.Warn("could not extract user uuid from bearer token")
		h.sendErrorWithCORS(w, http.StatusBadRequest, "Invalid bearer token")
		return
	}

	// Get user to find their wallet address
	user, err := h.store.GetUser(userUUID)
	if err != nil || user == nil {
		logger.Error("user not found", zap.String("user_id", userUUID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusNotFound, "User not found")
		return
	}

	// Get user's wallets to find the deposit address
	wallets, err := h.store.GetWalletsByUser(userUUID)
	if err != nil {
		logger.Error("failed to get wallets for user", zap.String("user_id", userUUID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to get wallets")
		return
	}

	// If no wallets exist, create one automatically (consistent with GetUserWallets)
	if len(wallets) == 0 {
		logger.Info("no wallets found for user, creating one automatically", zap.String("user_id", userUUID))
		address := utils.GenerateMockXRPLAddress()
		wallet := &models.Wallet{
			Address: address,
			UserID:  userUUID,
			Name:    "Default Wallet",
			Type:    consts.WalletTypeStandard,
			Network: consts.NetworkXRPLedger,
		}
		if err := h.store.CreateWallet(wallet); err != nil {
			logger.Error("failed to auto-create wallet", zap.String("user_id", userUUID), zap.Error(err))
			h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to create wallet")
			return
		}
		wallets = []*models.Wallet{wallet}
	}

	// Use the first wallet's address
	walletAddress := wallets[0].Address

	// Get vault_uuid for the currency
	vaultUUID := consts.SandboxVaultIDs[txReq.Currency]

	// Parse amount as float
	amountFloat, err := strconv.ParseFloat(txReq.Amount, 64)
	if err != nil {
		logger.Error("invalid amount format in deposit request", zap.String("amount", txReq.Amount), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusBadRequest, "Invalid amount format")
		return
	}
	amountStr := formatAmount(amountFloat)

	// Calculate deposit fee
	feePercent, _ := h.feeConfig.GetDepositFeeForUser(userUUID)
	feeAmount := CalculateFee(amountFloat, feePercent)
	feeStr := formatAmount(feeAmount)
	// For deposits, total_amount = amount (fee is charged separately by GateHub)
	totalAmountStr := amountStr

	txID := utils.GenerateUUID()

	tx := &models.Transaction{
		ID:               txID,
		UserID:           userUUID,
		Amount:           amountStr,
		TotalAmount:      totalAmountStr,
		Fee:              feeStr,
		Currency:         txReq.Currency,
		VaultUUID:        vaultUUID,
		ReceivingAddress: walletAddress,
		Type:             consts.TransactionTypeDeposit,
		DepositType:      consts.DepositTypeExternal,
		Status:           1, // 1 = completed
	}

	if err := h.store.CreateTransaction(tx); err != nil {
		logger.Error("failed to create transaction", zap.String("transaction_id", txID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	if err := h.store.AddBalance(userUUID, txReq.Currency, amountFloat); err != nil {
		logger.Error("failed to update balance for user", zap.String("user_id", userUUID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to update balance")
		return
	}

	// Send deposit webhook (matches GateHub webhook spec) with dynamic values
	h.webhookManager.SendAsync("core.deposit.completed", userUUID, map[string]interface{}{
		"tx_uuid":      txID,
		"amount":       amountStr,      // From iframe form
		"currency":     txReq.Currency, // From iframe form
		"vault_uuid":   vaultUUID,      // Vault UUID for this currency
		"address":      walletAddress,  // The wallet address that received the deposit
		"deposit_type": "external",     // External deposit type (lowercase per spec)
		"total_fees":   "0",            // Fees charged (matches GateHub spec)
	}, 0)

	logger.Info("sent deposit webhook", zap.String("user_id", userUUID), zap.String("amount", amountStr), zap.String("currency", txReq.Currency), zap.String("wallet_address", walletAddress))

	// Return success response
	h.sendJSONWithCORS(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Transaction completed",
	})
}

// processWithdrawal handles the withdrawal transaction logic
func (h *Handler) processWithdrawal(w http.ResponseWriter, bearer string, txReq *TransactionRequest) {
	// Decode bearer to get user UUID
	userUUID := h.extractUserFromBearer(bearer)

	if userUUID == "" {
		logger.Warn("could not extract user uuid from bearer token")
		h.sendErrorWithCORS(w, http.StatusBadRequest, "Invalid bearer token")
		return
	}

	// Get user to find their wallet address
	user, err := h.store.GetUser(userUUID)
	if err != nil || user == nil {
		logger.Error("user not found", zap.String("user_id", userUUID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusNotFound, "User not found")
		return
	}

	// Get user's wallets to find the withdrawal source
	wallets, err := h.store.GetWalletsByUser(userUUID)
	if err != nil {
		logger.Error("failed to get wallets for user", zap.String("user_id", userUUID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to get wallets")
		return
	}

	if len(wallets) == 0 {
		logger.Error("no wallets found for user", zap.String("user_id", userUUID))
		h.sendErrorWithCORS(w, http.StatusBadRequest, "User has no wallets")
		return
	}

	// Use the first wallet's address
	walletAddress := wallets[0].Address

	// Get vault_uuid for the currency
	vaultUUID := consts.SandboxVaultIDs[txReq.Currency]

	// Parse amount as float
	amountFloat, err := strconv.ParseFloat(txReq.Amount, 64)
	if err != nil {
		logger.Error("invalid amount format in withdrawal request", zap.String("amount", txReq.Amount), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusBadRequest, "Invalid amount format")
		return
	}
	amountStr := formatAmount(amountFloat)

	// Calculate withdrawal fee
	feePercent, _ := h.feeConfig.GetWithdrawalFeeForUser(userUUID)
	feeAmount := CalculateFee(amountFloat, feePercent)
	feeStr := formatAmount(feeAmount)

	// For withdrawals, the amount is deducted (total_amount includes the fee deducted)
	totalAmount := amountFloat + feeAmount // Total deducted from user's balance
	totalAmountStr := formatAmount(totalAmount)

	// Check if user has sufficient balance
	currentBalance, _ := h.store.GetBalance(userUUID, txReq.Currency)
	if currentBalance < totalAmount {
		logger.Warn("insufficient balance for withdrawal",
			zap.String("user_id", userUUID),
			zap.String("currency", txReq.Currency),
			zap.Float64("requested_total", totalAmount),
			zap.Float64("current_balance", currentBalance))
		h.sendErrorWithCORS(w, http.StatusBadRequest, fmt.Sprintf("Insufficient balance. Required: %.2f, Available: %.2f", totalAmount, currentBalance))
		return
	}

	txID := utils.GenerateUUID()

	tx := &models.Transaction{
		ID:               txID,
		UserID:           userUUID,
		Amount:           amountStr,
		TotalAmount:      totalAmountStr,
		Fee:              feeStr,
		Currency:         txReq.Currency,
		VaultUUID:        vaultUUID,
		ReceivingAddress: walletAddress,
		Type:             consts.TransactionTypeWithdrawal,
		DepositType:      consts.DepositTypeWithdrawal,
		Status:           consts.TransactionStatusPending,
		AccountIBAN:      "GB29NWBK60161331926819",
		AccountLegalName: "Jane Smith",
		Message:          "Mock Reference",
	}

	if err := h.store.CreateTransaction(tx); err != nil {
		logger.Error("failed to create withdrawal transaction", zap.String("transaction_id", txID), zap.Error(err))
		h.sendErrorWithCORS(w, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	logger.Info("withdrawal created as pending", zap.String("user_id", userUUID), zap.String("transaction_id", txID), zap.String("amount", amountStr), zap.String("currency", txReq.Currency))

	// Return success response with transaction ID for iframe
	h.sendJSONWithCORS(w, http.StatusOK, map[string]string{
		"status":         "success",
		"message":        "Withdrawal pending",
		"transaction_id": txID, // Frontend needs this for CreateGatehubWithdrawal
		"uuid":           txID, // Alias for compatibility
	})
}

// It looks up the token in the stored token->user UUID mapping
func (h *Handler) extractUserFromBearer(bearer string) string {
	// Look up the user UUID from the token mapping
	if userUUID, ok := h.tokenToUser.Load(bearer); ok {
		if uuid, ok := userUUID.(string); ok {
			return uuid
		}
	}

	logger.Debug("bearer token not found in mapping", zap.String("token_prefix", tokenPrefix(bearer)))
	return ""
}

// extractBearerFromAuthHeader extracts the bearer token from the Authorization header
func extractBearerFromAuthHeader(authHeader string) string {
	if len(authHeader) < 7 {
		return ""
	}

	if authHeader[:7] != "Bearer " {
		return ""
	}

	return authHeader[7:] // Return everything after "Bearer "
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ListWithdrawals returns withdrawals for a user, optionally filtered by status.
// GET /admin/users/{userID}/withdrawals?status=pending
func (h *Handler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	statusFilter := r.URL.Query().Get("status")

	txs, err := h.store.ListTransactionsByUser(userID)
	if err != nil {
		logger.Error("admin: failed to list transactions", zap.String("user_id", userID), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	var withdrawals []*models.Transaction
	for _, tx := range txs {
		if tx.DepositType != consts.DepositTypeWithdrawal {
			continue
		}
		if statusFilter == "pending" && tx.Status != consts.TransactionStatusPending {
			continue
		}
		withdrawals = append(withdrawals, tx)
	}

	if withdrawals == nil {
		withdrawals = []*models.Transaction{}
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].CreatedAt.After(withdrawals[j].CreatedAt)
	})

	h.sendJSON(w, http.StatusOK, withdrawals)
}

// TriggerWithdrawalEvent fires a withdrawal webhook for a pending transaction.
// POST /admin/withdrawals/{txID}/trigger-event
// Body: {"event": "core.withdrawal.completed" | "more-bridge.withdrawal.rejected"}
func (h *Handler) TriggerWithdrawalEvent(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")

	var req struct {
		Event string `json:"event"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Event {
	case consts.WebhookEventWithdrawalCompleted, consts.WebhookEventWithdrawalRejected:
	default:
		h.sendError(w, http.StatusBadRequest, "unsupported event; use core.withdrawal.completed or more-bridge.withdrawal.rejected")
		return
	}

	tx, err := h.store.GetTransaction(txID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "transaction not found")
		return
	}

	if tx.DepositType != consts.DepositTypeWithdrawal {
		h.sendError(w, http.StatusBadRequest, "transaction is not a withdrawal")
		return
	}

	if tx.Status != consts.TransactionStatusPending {
		h.sendError(w, http.StatusConflict, "transaction is not pending")
		return
	}

	switch req.Event {
	case consts.WebhookEventWithdrawalCompleted:
		totalAmount, err := strconv.ParseFloat(tx.TotalAmount, 64)
		if err != nil {
			logger.Error("admin: failed to parse total amount", zap.String("tx_id", txID), zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "failed to parse transaction amount")
			return
		}
		if err := h.store.DeductBalance(tx.UserID, tx.Currency, totalAmount); err != nil {
			logger.Error("admin: failed to deduct balance", zap.String("tx_id", txID), zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "failed to deduct balance")
			return
		}
		if err := h.store.UpdateTransactionStatus(txID, consts.TransactionStatusCompleted); err != nil {
			logger.Error("admin: failed to update transaction status", zap.String("tx_id", txID), zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "failed to update transaction status")
			return
		}
		h.webhookManager.SendAsync(consts.WebhookEventWithdrawalCompleted, tx.UserID, map[string]interface{}{
			"tx_uuid":    txID,
			"amount":     tx.Amount,
			"currency":   tx.Currency,
			"address":    tx.ReceivingAddress,
			"total_fees": tx.Fee,
		}, 0)
		logger.Info("admin: withdrawal completed", zap.String("tx_id", txID), zap.String("user_id", tx.UserID))

	case consts.WebhookEventWithdrawalRejected:
		if err := h.store.UpdateTransactionStatus(txID, consts.TransactionStatusFailed); err != nil {
			logger.Error("admin: failed to update transaction status", zap.String("tx_id", txID), zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "failed to update transaction status")
			return
		}
		h.webhookManager.SendAsync(consts.WebhookEventWithdrawalRejected, tx.UserID, map[string]interface{}{
			"txUuid":   txID,
			"userUuid": tx.UserID,
			"status":   "PENDING",
			"currency": tx.Currency,
			"amount":   tx.Amount,
		}, 0)
		logger.Info("admin: withdrawal rejected", zap.String("tx_id", txID), zap.String("user_id", tx.UserID))
	}

	h.sendJSON(w, http.StatusOK, map[string]string{
		"tx_id": txID,
		"event": req.Event,
	})
}
