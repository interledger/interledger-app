package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
)

type createWalletRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
}

type walletVerificationResponse struct {
	Status string `json:"status"`
}

type walletResponse struct {
	ID           string                     `json:"id"`
	Parent       string                     `json:"parent"`
	UID          string                     `json:"uid"`
	Name         string                     `json:"name"`
	SubAccount   bool                       `json:"subAccount"`
	Verification walletVerificationResponse `json:"verification"`
}

// CreateWallet handles Chimoney sub-account creation.
func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req createWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		h.respondErr(w, http.StatusBadRequest, "name is required")
		return
	}

	account := models.SubAccount{
		ID:          uuid.NewString(),
		ParentID:    "mockchimoney-root",
		UID:         uuid.NewString(),
		Name:        req.Name,
		Email:       strings.TrimSpace(req.Email),
		FirstName:   strings.TrimSpace(req.FirstName),
		LastName:    strings.TrimSpace(req.LastName),
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
		SubAccount:  true,
		KYCStatus:   "pending",
		CreatedAt:   time.Now().UTC(),
	}

	created, err := h.store.CreateSubAccount(r.Context(), account)
	if err != nil {
		h.respondErr(w, http.StatusInternalServerError, "failed to create wallet")
		return
	}

	h.respondOK(w, http.StatusCreated, toWalletResponse(created))
}

// GetWallet returns a Chimoney sub-account by ID.
func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		h.respondErr(w, http.StatusBadRequest, "id is required")
		return
	}

	account, err := h.store.GetSubAccount(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondErr(w, http.StatusNotFound, "wallet not found")
			return
		}
		h.respondErr(w, http.StatusInternalServerError, "failed to get wallet")
		return
	}

	h.respondOK(w, http.StatusOK, toWalletResponse(account))
}

func toWalletResponse(account models.SubAccount) walletResponse {
	return walletResponse{
		ID:         account.ID,
		Parent:     account.ParentID,
		UID:        account.UID,
		Name:       account.Name,
		SubAccount: account.SubAccount,
		Verification: walletVerificationResponse{
			Status: account.KYCStatus,
		},
	}
}
