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

	beneficiaries, total, err := h.store.ListBeneficiariesByWallet(r.Context(), subAcc.AccountID, limit, offset)
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
