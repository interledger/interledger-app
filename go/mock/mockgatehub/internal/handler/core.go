package handler

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// resolveWalletParam returns the wallet ID from URL params, trying "walletID"
// first and falling back to the legacy "address" parameter name.
func resolveWalletParam(r *http.Request) string {
	if id := chi.URLParam(r, "walletID"); id != "" {
		return id
	}
	return chi.URLParam(r, "address")
}

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWalletRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get userID from path parameter if not in body
	userID := chi.URLParam(r, "userID")
	if req.UserID == "" && userID != "" {
		req.UserID = userID
	}

	if req.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	logger.Info("creating wallet", zap.String("user_id", req.UserID))

	_, err := h.store.GetUser(req.UserID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	address := utils.GenerateMockXRPLAddress()

	if req.Type == 0 {
		req.Type = consts.WalletTypeStandard
	}
	if req.Network == 0 {
		req.Network = consts.NetworkXRPLedger
	}

	wallet := &models.Wallet{
		Address: address,
		UserID:  req.UserID,
		Name:    req.Name,
		Type:    req.Type,
		Network: req.Network,
	}

	if err := h.store.CreateWallet(wallet); err != nil {
		logger.Error("failed to create wallet", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to create wallet")
		return
	}

	logger.Info("wallet created", zap.String("address", address), zap.String("user_id", req.UserID))
	h.sendJSON(w, http.StatusCreated, wallet)
}

// GetUserWallets retrieves all wallets for a user (GET /core/v1/users/{userID})
// If no wallets exist, creates one automatically
func (h *Handler) GetUserWallets(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	logger.Info("getting wallets for user", zap.String("user_id", userID))

	_, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	wallets, err := h.store.GetWalletsByUser(userID)
	if err != nil {
		logger.Error("failed to get user wallets", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to get wallets")
		return
	}

	// If no wallets exist, create one automatically
	if len(wallets) == 0 {
		logger.Info("no wallets found for user, creating one automatically", zap.String("user_id", userID))
		address := utils.GenerateMockXRPLAddress()
		wallet := &models.Wallet{
			Address: address,
			UserID:  userID,
			Name:    "Default Wallet",
			Type:    consts.WalletTypeStandard,
			Network: consts.NetworkXRPLedger,
		}

		if err := h.store.CreateWallet(wallet); err != nil {
			logger.Error("failed to create wallet", zap.Error(err))
			h.sendError(w, http.StatusInternalServerError, "Failed to create wallet")
			return
		}

		logger.Info("created default wallet", zap.String("address", address), zap.String("user_id", userID))
		wallets = append(wallets, wallet)
	}

	// Return in the format expected by wallet-backend: { wallets: [...] }
	response := models.UserWalletsResponse{}
	response.Wallets = make([]models.UserWalletResponse, 0, len(wallets))

	for i, w := range wallets {
		response.Wallets = append(response.Wallets, models.UserWalletResponse{
			UUID:    w.Address,
			Address: w.Address,
			Name:    w.Name,
			Type:    w.Type,
			Primary: i == 0,
			Active:  true,
			Enabled: true,
		})
	}

	logger.Info("returning wallets", zap.Int("count", len(wallets)), zap.String("user_id", userID))
	h.sendJSON(w, http.StatusOK, response)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	walletID := resolveWalletParam(r)
	if walletID == "" {
		h.sendError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	logger.Info("getting wallet", zap.String("wallet_id", walletID))

	wallet, err := h.store.GetWallet(walletID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Wallet not found")
		return
	}

	h.sendJSON(w, http.StatusOK, wallet)
}

func (h *Handler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	walletID := resolveWalletParam(r)
	if walletID == "" {
		h.sendError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	logger.Info("getting balance for wallet", zap.String("wallet_id", walletID))

	wallet, err := h.store.GetWallet(walletID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Wallet not found")
		return
	}

	balances := make([]models.WalletBalanceResponse, 0, len(consts.SandboxCurrencies))
	for _, currency := range consts.SandboxCurrencies {
		balance, _ := h.store.GetBalance(wallet.UserID, currency)
		balances = append(balances, models.WalletBalanceResponse{
			Available: formatAmount(balance),
			Pending:   "0.00",
			Total:     formatAmount(balance),
			Vault: models.VaultSummary{
				UUID:      consts.SandboxVaultIDs[currency],
				Name:      fmt.Sprintf("Sandbox Vault %s", currency),
				AssetCode: currency,
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
		})
	}

	logger.Info("returning currency balances", zap.Int("count", len(balances)), zap.String("wallet_id", walletID))

	h.sendJSON(w, http.StatusOK, balances)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If user_id not in body, try to get from x-gatehub-managed-user-uuid header
	if req.UserID == "" {
		req.UserID = r.Header.Get("x-gatehub-managed-user-uuid")
		logger.Info("transaction: attempting to extract user_id from header", zap.String("user_id", req.UserID))
	}

	// If still no user_id, try to look up from receiving_address (wallet)
	if req.UserID == "" && req.ReceivingAddress != "" {
		wallet, err := h.store.GetWallet(req.ReceivingAddress)
		if err == nil && wallet != nil {
			req.UserID = wallet.UserID
			logger.Info("transaction: resolved user_id from receiving_address", zap.String("user_id", req.UserID), zap.String("receiving_address", req.ReceivingAddress))
		}
	}

	if req.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "user_id is required (provide in body, x-gatehub-managed-user-uuid header, or use a valid receiving_address)")
		return
	}
	if req.Amount <= 0 {
		h.sendError(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	// Infer currency from vault_uuid if not provided
	if req.Currency == "" {
		if req.VaultUUID == "" {
			h.sendError(w, http.StatusBadRequest, "either currency or vault_uuid is required")
			return
		}
		// Look up currency from vault_uuid
		currency, exists := consts.VaultUUIDToCurrency[req.VaultUUID]
		if !exists {
			h.sendError(w, http.StatusBadRequest, "invalid vault_uuid")
			return
		}
		req.Currency = currency
		logger.Info("inferred currency from vault uuid", zap.String("currency", currency), zap.String("vault_uuid", req.VaultUUID))
	} else {
		// If currency is provided, ensure vault_uuid matches (if also provided)
		if req.VaultUUID != "" {
			expectedVaultUUID := consts.SandboxVaultIDs[req.Currency]
			if req.VaultUUID != expectedVaultUUID {
				// Vault UUID mismatch: got vault, expected vault for currency. Using vault_uuid to determine currency
				// Trust vault_uuid over currency parameter
				if inferredCurrency, exists := consts.VaultUUIDToCurrency[req.VaultUUID]; exists {
					req.Currency = inferredCurrency
					// Corrected currency based on vault_uuid
				}
			}
		}
	}

	// Creating transaction
	_, err := h.store.GetUser(req.UserID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	if req.Type == 0 {
		req.Type = consts.TransactionTypeDeposit
	}
	if req.DepositType == "" {
		if req.Type == consts.TransactionTypeDeposit {
			req.DepositType = consts.DepositTypeExternal
		} else {
			req.DepositType = consts.DepositTypeHosted
		}
	}

	// Ensure vault_uuid is set based on currency
	if req.VaultUUID == "" {
		req.VaultUUID = consts.SandboxVaultIDs[req.Currency]
	}

	// Format amounts as strings to match GateHub API
	amountStr := formatAmount(req.Amount)

	// Calculate fee based on transaction type:
	// - External deposits: use deposit fee percentage
	// - Withdrawals: use withdrawal fee percentage
	// - Hosted transfers: always free
	var feePercent float64
	switch req.DepositType {
	case consts.DepositTypeExternal:
		feePercent, _ = h.feeConfig.GetDepositFeeForUser(req.UserID)
	case "withdrawal":
		feePercent, _ = h.feeConfig.GetWithdrawalFeeForUser(req.UserID)
	}
	feeAmount := CalculateFee(req.Amount, feePercent)
	feeStr := formatAmount(feeAmount)
	// For deposits: total_amount = amount (fee is charged separately by GateHub)
	// For withdrawals: total_amount = amount + fee (total deducted)
	totalAmountStr := amountStr
	if req.DepositType == "withdrawal" {
		totalAmountStr = formatAmount(req.Amount + feeAmount)
	}

	tx := &models.Transaction{
		UserID:           req.UserID,
		UID:              req.UID,
		Amount:           amountStr,
		TotalAmount:      totalAmountStr,
		Fee:              feeStr,
		Currency:         req.Currency,
		VaultUUID:        req.VaultUUID,
		SendingAddress:   req.SendingAddress,
		ReceivingAddress: req.ReceivingAddress,
		Type:             req.Type,
		DepositType:      req.DepositType,
		Status:           consts.TransactionStatusPending,
	}

	if err := h.store.CreateTransaction(tx); err != nil {
		logger.Error("failed to create transaction", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to create transaction")
		return
	}

	// Simulate real-world transaction processing:
	// 1) Immediately emit a PENDING webhook
	// 2) After a short delay, mark completed, emit COMPLETED webhook, and update balance
	// This keeps Temporal workflows from hanging on long polling timers.
	respTx := *tx // snapshot to avoid mutating response status during async completion

	if req.DepositType == consts.DepositTypeExternal || req.DepositType == consts.DepositTypeHosted {
		txID := tx.ID
		userID := req.UserID
		currency := req.Currency
		amount := req.Amount
		feeAmount := feeAmount // Capture fee amount from outer scope
		depositType := req.DepositType
		sendingAddr := req.SendingAddress
		receivingAddr := req.ReceivingAddress
		hasWebhook := h.webhookManager.HasURL()

		// Determine balance direction for hosted transfers.
		// If the sending_address is a known user wallet, money is leaving the user → debit.
		// Otherwise (receiving_address is a known wallet), money is arriving → credit.
		isDebit := false
		if depositType == consts.DepositTypeHosted && sendingAddr != "" {
			if wallet, err := h.store.GetWallet(sendingAddr); err == nil && wallet != nil && wallet.UserID == userID {
				isDebit = true
			}
		}

		pendingPayload := map[string]interface{}{
			"transaction_id": txID,
			"tx_uuid":        txID,
			"amount":         tx.Amount,
			"currency":       tx.Currency,
			"address":        receivingAddr,
			"deposit_type":   depositType,
			"status":         "pending",
		}
		h.webhookManager.SendAsync(consts.WebhookEventDepositCompleted, userID, pendingPayload, 0)

		complete := func() {
			if hasWebhook {
				if err := h.store.UpdateTransactionStatus(txID, consts.TransactionStatusCompleted); err != nil {
					logger.Error("failed to update transaction to completed", zap.String("transaction_id", txID), zap.Error(err))
					return
				}
			}

			markFailed := func() {
				if !hasWebhook {
					return
				}
				if statusErr := h.store.UpdateTransactionStatus(txID, consts.TransactionStatusFailed); statusErr != nil {
					logger.Error("failed to mark transaction as failed", zap.String("transaction_id", txID), zap.Error(statusErr))
				}
				failedPayload := map[string]interface{}{
					"transaction_id": txID,
					"tx_uuid":        txID,
					"amount":         formatAmount(amount),
					"currency":       currency,
					"address":        receivingAddr,
					"deposit_type":   depositType,
					"status":         "failed",
				}
				h.webhookManager.SendAsync(consts.WebhookEventDepositCompleted, userID, failedPayload, 0)
			}

			netAmount := amount - feeAmount
			if isDebit {
				// Hosted transfer where user is the sender → debit
				if err := h.store.DeductBalance(userID, currency, netAmount); err != nil {
					logger.Error("failed to deduct balance for transaction", zap.String("transaction_id", txID), zap.Error(err))
					markFailed()
					return
				}
				logger.Info("transaction completed (debit)",
					zap.String("transaction_id", txID),
					zap.Float64("amount", amount),
					zap.Float64("fee", feeAmount),
					zap.Float64("net_amount", netAmount),
					zap.String("currency", currency))
			} else {
				// Deposit or hosted transfer where user is the receiver → credit
				if err := h.store.AddBalance(userID, currency, netAmount); err != nil {
					logger.Error("failed to update balance for transaction", zap.String("transaction_id", txID), zap.Error(err))
					markFailed()
					return
				}
				logger.Info("transaction completed (credit)",
					zap.String("transaction_id", txID),
					zap.Float64("amount", amount),
					zap.Float64("fee", feeAmount),
					zap.Float64("net_amount", netAmount),
					zap.String("currency", currency))
			}

			if hasWebhook {
				completedPayload := map[string]interface{}{
					"transaction_id": txID,
					"tx_uuid":        txID,
					"amount":         formatAmount(amount),
					"currency":       currency,
					"address":        receivingAddr,
					"deposit_type":   depositType,
					"status":         "completed",
				}
				h.webhookManager.SendAsync(consts.WebhookEventDepositCompleted, userID, completedPayload, 0)
			}
		}

		if !hasWebhook {
			// For test/local runs without a webhook target, complete immediately to satisfy balance expectations
			complete()
		} else {
			delay := 2 * time.Second
			go func() {
				if delay > 0 {
					time.Sleep(delay)
				}
				complete()
			}()
		}
	}

	h.sendJSON(w, http.StatusCreated, &respTx)
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")
	if txID == "" {
		h.sendError(w, http.StatusBadRequest, "Transaction ID is required")
		return
	}

	logger.Info("getting transaction", zap.String("transaction_id", txID))

	tx, err := h.store.GetTransaction(txID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Transaction not found")
		return
	}

	// Transform to match GateHub API format with nested objects
	response := map[string]interface{}{
		"uuid":         tx.ID,
		"user_id":      tx.UserID,
		"amount":       tx.Amount,
		"total_amount": tx.TotalAmount,
		"fee":          tx.Fee,
		"currency":     tx.Currency,
		"type":         tx.Type,
		"deposit_type": tx.DepositType,
		"status":       tx.Status,
		"created_at":   tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"vault": map[string]interface{}{
			"uuid":       tx.VaultUUID,
			"asset_code": tx.Currency,
			"name":       tx.Currency + " Vault",
		},
		"sending_wallet": map[string]interface{}{
			"address": tx.ReceivingAddress, // For withdrawals, this is the sending wallet
		},
		"receiving_wallet": map[string]interface{}{
			"address": tx.ReceivingAddress,
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetUserCurrencies returns the list of currencies the user has accounts for
func (h *Handler) GetUserCurrencies(w http.ResponseWriter, r *http.Request) {
	bearer := r.URL.Query().Get("bearer")
	if bearer == "" {
		bearer = r.Header.Get("Authorization")
		if len(bearer) > 7 && bearer[:7] == "Bearer " {
			bearer = bearer[7:]
		}
	}

	if bearer == "" {
		h.sendError(w, http.StatusBadRequest, "Missing bearer token")
		return
	}

	// Extract user UUID from bearer token
	userUUID := h.extractUserFromBearer(bearer)
	if userUUID == "" {
		// Return default currencies if we can't determine user
		logger.Debug("could not extract user from bearer, returning all currencies")
		h.sendJSON(w, http.StatusOK, models.CurrenciesResponse{
			Currencies: consts.SandboxCurrencies,
		})
		return
	}

	logger.Info("getting currencies for user", zap.String("user_id", userUUID))

	// Get currencies that have non-zero balances for this user
	userCurrencies := []string{}

	for _, currency := range consts.SandboxCurrencies {
		balance, err := h.store.GetBalance(userUUID, currency)
		if err == nil && balance > 0 {
			userCurrencies = append(userCurrencies, currency)
		}
	}

	// If no currencies with balance, return all currencies (user hasn't deposited yet)
	if len(userCurrencies) == 0 {
		logger.Info("no balances found for user, returning all currencies", zap.String("user_id", userUUID))
		h.sendJSON(w, http.StatusOK, models.CurrenciesResponse{
			Currencies: consts.SandboxCurrencies,
		})
		return
	}

	logger.Info("found currencies with balances", zap.String("user_id", userUUID), zap.Int("count", len(userCurrencies)), zap.Strings("currencies", userCurrencies))

	h.sendJSON(w, http.StatusOK, models.CurrenciesResponse{
		Currencies: userCurrencies,
	})
}
