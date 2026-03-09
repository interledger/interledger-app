package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
	"gitlab.com/fynbos/mock/mockxago/internal/utils"
)

// AddBeneficiary handles POST /v1/accounts/{accountId}/beneficiaries
func (h *Handler) AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "account ID is required")
		return
	}

	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "sub-account not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load sub-account")
		return
	}

	var req models.AddBeneficiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if strings.TrimSpace(req.AccountNumber) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "accountNumber is required")
		return
	}

	beneficiary := &models.Beneficiary{
		ID:        utils.GenerateUUID(),
		AccountID: accountID,
		// Store the actual WalletID so that transfers can properly look up balances
		// Beneficiary isolation by account is handled at the storage layer
		WalletID:      subAcc.WalletID,
		Name:          req.Name,
		Scope:         req.Scope,
		IsOwn:         req.IsOwn,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		BranchCode:    req.BranchCode,
		Reference:     req.Reference,
		Currency:      req.CurrencyCode,
		Status:        "pending",
	}

	if err := h.store.SaveBeneficiary(r.Context(), beneficiary); err != nil {
		logger.Errorf("Failed to save beneficiary: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to save beneficiary")
		return
	}

	// Auto-approve after 3 seconds (sandbox behaviour), use a fresh context
	go func(id string) {
		time.Sleep(3 * time.Second)
		if err := h.store.UpdateBeneficiaryStatus(context.Background(), id, "approved"); err != nil {
			logger.Warnf("Failed to auto-approve beneficiary %s: %v", id, err)
		} else {
			logger.Infof("Auto-approved beneficiary %s", id)
		}
	}(beneficiary.ID)

	logger.Infof("Created beneficiary %s for account %s (wallet %s)", beneficiary.ID, accountID, subAcc.WalletID)

	h.sendJSON(w, http.StatusOK, beneficiaryToItem(beneficiary))
}

// ListBeneficiaries handles GET /v1/accounts/{accountId}/beneficiaries
func (h *Handler) ListBeneficiaries(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "account ID is required")
		return
	}

	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "sub-account not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load sub-account")
		return
	}

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

	beneficiaries, total, err := h.store.ListBeneficiariesByAccountID(r.Context(), subAcc.AccountID, limit, offset)
	if err != nil {
		logger.Errorf("Failed to list beneficiaries: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to list beneficiaries")
		return
	}

	numberOfPages := 1
	if total > 0 {
		numberOfPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	items := make([]models.BeneficiaryItem, 0, len(beneficiaries))
	for _, b := range beneficiaries {
		items = append(items, beneficiaryToItem(b))
	}

	resp := models.ListBeneficiariesResponse{
		Data: items,
		Pagination: models.BeneficiaryPagination{
			Limit:         limit,
			Page:          page,
			NumberOfPages: numberOfPages,
			Total:         total,
		},
	}

	logger.Infof("Listed %d beneficiaries for account %s (page=%d, limit=%d, total=%d)", len(items), accountID, page, limit, total)
	h.sendJSON(w, http.StatusOK, resp)
}

// AddBeneficiaryGlobal handles POST /v1/beneficiaries (without accountId in path)
// Resolves accountId from the bearer token context
func (h *Handler) AddBeneficiaryGlobal(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
		return
	}

	// Parse "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
		return
	}
	tokenValue := parts[1]

	// Resolve accountId from token
	accountID, err := h.store.GetAccountIDByToken(r.Context(), tokenValue)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "no account associated with token")
		return
	}

	// Validate sub-account exists
	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "sub-account not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load sub-account")
		return
	}

	var req models.AddBeneficiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	if strings.TrimSpace(req.AccountNumber) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "accountNumber is required")
		return
	}

	beneficiary := &models.Beneficiary{
		ID:            utils.GenerateUUID(),
		AccountID:     accountID,
		WalletID:      subAcc.WalletID,
		Name:          req.Name,
		Scope:         req.Scope,
		IsOwn:         req.IsOwn,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		BranchCode:    req.BranchCode,
		Reference:     req.Reference,
		Currency:      req.CurrencyCode,
		Status:        "pending",
	}

	if err := h.store.SaveBeneficiary(r.Context(), beneficiary); err != nil {
		logger.Errorf("Failed to save beneficiary: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to save beneficiary")
		return
	}

	// Auto-approve after 3 seconds (sandbox behaviour)
	go func(id string) {
		time.Sleep(3 * time.Second)
		if err := h.store.UpdateBeneficiaryStatus(context.Background(), id, "approved"); err != nil {
			logger.Warnf("Failed to auto-approve beneficiary %s: %v", id, err)
		} else {
			logger.Infof("Auto-approved beneficiary %s", id)
		}
	}(beneficiary.ID)

	logger.Infof("Created beneficiary %s for account %s via global endpoint", beneficiary.ID, accountID)

	h.sendJSON(w, http.StatusOK, beneficiaryToItem(beneficiary))
}

// ListBeneficiariesGlobal handles GET /v1/beneficiaries (without accountId in path)
// Resolves accountId from the bearer token context
func (h *Handler) ListBeneficiariesGlobal(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
		return
	}

	// Parse "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
		return
	}
	tokenValue := parts[1]

	// Resolve accountId from token
	accountID, err := h.store.GetAccountIDByToken(r.Context(), tokenValue)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "no account associated with token")
		return
	}

	// Validate sub-account exists
	subAcc, err := h.store.GetSubAccount(r.Context(), accountID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "sub-account not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load sub-account")
		return
	}

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

	beneficiaries, total, err := h.store.ListBeneficiariesByAccountID(r.Context(), subAcc.AccountID, limit, offset)
	if err != nil {
		logger.Errorf("Failed to list beneficiaries: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to list beneficiaries")
		return
	}

	numberOfPages := 1
	if total > 0 {
		numberOfPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	items := make([]models.BeneficiaryItem, 0, len(beneficiaries))
	for _, b := range beneficiaries {
		items = append(items, beneficiaryToItem(b))
	}

	resp := models.ListBeneficiariesResponse{
		Data: items,
		Pagination: models.BeneficiaryPagination{
			Limit:         limit,
			Page:          page,
			NumberOfPages: numberOfPages,
			Total:         total,
		},
	}

	logger.Infof("Listed %d beneficiaries for account %s via global endpoint (page=%d, limit=%d, total=%d)", len(items), accountID, page, limit, total)
	h.sendJSON(w, http.StatusOK, resp)
}

func beneficiaryToItem(b *models.Beneficiary) models.BeneficiaryItem {
	return models.BeneficiaryItem{
		UUID:          b.ID,
		Name:          b.Name,
		Scope:         b.Scope,
		CurrencyCode:  b.Currency,
		AccountNumber: b.AccountNumber,
		BranchCode:    b.BranchCode,
		BankName:      b.BankName,
		AccountName:   b.AccountName,
		Reference:     b.Reference,
		IsOwn:         b.IsOwn,
		Status:        b.Status,
	}
}
