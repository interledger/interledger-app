package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
	"gitlab.com/fynbos/mock/mockpti/internal/utils"
)

// CreatePaymentInformation handles POST /users/{id}/payment-information.
func (h *Handler) CreatePaymentInformation(w http.ResponseWriter, r *http.Request) {
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

	var req models.CreatePaymentInformationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	piID := utils.GenerateUUID()

	pi := &models.PaymentInformation{
		ID:                    piID,
		Type:                  req.Type,
		BankAccountNumber:     req.BankAccountNumber,
		BankAccountType:       req.BankAccountType,
		BankSwiftCode:         req.BankSwiftCode,
		BankRoutingNumber:     req.BankRoutingNumber,
		BankRoutingCheckDigit: req.BankRoutingCheckDigit,
		AccountBankName:       req.AccountBankName,
		UserID:                userID,
	}

	if err := h.store.SavePaymentInformation(r.Context(), pi); err != nil {
		logger.Errorf("failed to save payment information: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create payment information")
		return
	}

	logger.Infof("Created payment information %s for user %s", piID, userID)

	h.sendJSON(w, http.StatusOK, pi)
}

// GetPaymentInformation handles GET /users/{id}/payment-information/{piId}.
func (h *Handler) GetPaymentInformation(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	piID := chi.URLParam(r, "piId")

	if userID == "" || piID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id and payment information id are required")
		return
	}

	pi, err := h.store.GetPaymentInformation(r.Context(), userID, piID)
	if err != nil {
		if errors.Is(err, storage.ErrPaymentInformationNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "payment information not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get payment information")
		return
	}

	h.sendJSON(w, http.StatusOK, pi)
}
