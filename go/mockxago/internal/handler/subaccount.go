package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
	"gitlab.com/fynbos/mockxago/internal/utils"
)

// CreateSubAccount handles POST /xago/v1/company/accounts
func (h *Handler) CreateSubAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if err := h.validateCreateSubAccount(req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	walletID := req.WalletID
	if strings.TrimSpace(walletID) == "" {
		walletID = utils.GenerateUUID()
	}

	subAcc := &models.SubAccount{
		ID:                        utils.GenerateUUID(),
		WalletID:                  walletID,
		AccountID:                 utils.GenerateUUID(),
		FirstName:                 req.FirstName,
		LastName:                  req.LastName,
		Email:                     req.Email,
		MobileNumber:              req.MobileNumber,
		IdentityType:              req.IdentityType,
		IDNumber:                  req.IDNumber,
		PhysicalAddress:           req.PhysicalAddress,
		ThirdPartyVerificationURL: req.ThirdPartyVerificationURL,
		DepositAddress:            generateDepositAddress(),
		DepositTag:                generateDepositTag(),
		DepositReferenceZAR:       fmt.Sprintf("%s_ZAR", walletID),
		DepositReferenceUSD:       fmt.Sprintf("%s_USD", walletID),
	}

	if err := h.store.SaveSubAccount(r.Context(), subAcc); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create sub-account")
		return
	}

	resp := models.CreateSubAccountResponse{
		AccountID:          subAcc.AccountID,
		DepositAddress:     subAcc.DepositAddress,
		DepositTag:         subAcc.DepositTag,
		BankDepositDetails: bankDepositDetails(),
		Beneficiaries: []models.BeneficiaryResponse{
			{
				BeneficiaryID:    utils.GenerateUUID(),
				BeneficiaryType:  "rollup",
				CurrencyID:       "ZAR",
				DepositReference: subAcc.DepositReferenceZAR,
				AccountNumber:    "62057334567",
				BankName:         "FNB",
				AccountName:      "Xago Holdings",
			},
			{
				BeneficiaryID:    utils.GenerateUUID(),
				BeneficiaryType:  "rollup",
				CurrencyID:       "USD",
				DepositReference: subAcc.DepositReferenceUSD,
				AccountNumber:    "0123456789",
				BankName:         "Citibank",
				AccountName:      "Xago Inc",
			},
		},
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// UpdateSubAccount handles PUT /xago/v1/company/accounts/{accountId}
func (h *Handler) UpdateSubAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if _, err := uuid.Parse(accountID); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid account ID format")
		return
	}

	var req models.UpdateSubAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
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

	if req.ThirdPartyVerificationURL != "" {
		subAcc.ThirdPartyVerificationURL = req.ThirdPartyVerificationURL
	}
	if req.IDNumber != "" {
		subAcc.IDNumber = req.IDNumber
	}
	if req.PhysicalAddress != "" {
		subAcc.PhysicalAddress = req.PhysicalAddress
	}

	if err := h.store.UpdateSubAccount(r.Context(), subAcc); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to update sub-account")
		return
	}

	h.sendJSON(w, http.StatusOK, models.UpdateSubAccountResponse{AccountID: accountID, Status: "updated"})
}

// GetSubAccountByWallet handles GET /xago/v1/company/accounts?walletId=...
func (h *Handler) GetSubAccountByWallet(w http.ResponseWriter, r *http.Request) {
	walletID := r.URL.Query().Get("walletId")
	if strings.TrimSpace(walletID) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "walletId is required")
		return
	}

	subAcc, err := h.store.GetSubAccountByWalletID(r.Context(), walletID)
	if err != nil {
		if err == storage.ErrSubAccountNotFound {
			h.sendError(w, http.StatusNotFound, "not_found", "sub-account not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to load sub-account")
		return
	}

	resp := models.CreateSubAccountResponse{
		AccountID:          subAcc.AccountID,
		DepositAddress:     subAcc.DepositAddress,
		DepositTag:         subAcc.DepositTag,
		BankDepositDetails: bankDepositDetails(),
		Beneficiaries: []models.BeneficiaryResponse{
			{
				BeneficiaryID:    utils.GenerateUUID(),
				BeneficiaryType:  "rollup",
				CurrencyID:       "ZAR",
				DepositReference: subAcc.DepositReferenceZAR,
				AccountNumber:    "62057334567",
				BankName:         "FNB",
				AccountName:      "Xago Holdings",
			},
			{
				BeneficiaryID:    utils.GenerateUUID(),
				BeneficiaryType:  "rollup",
				CurrencyID:       "USD",
				DepositReference: subAcc.DepositReferenceUSD,
				AccountNumber:    "0123456789",
				BankName:         "Citibank",
				AccountName:      "Xago Inc",
			},
		},
	}

	h.sendJSON(w, http.StatusOK, resp)
}

func (h *Handler) validateCreateSubAccount(req models.CreateSubAccountRequest) error {
	if strings.TrimSpace(req.FirstName) == "" {
		return fmt.Errorf("firstName is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return fmt.Errorf("lastName is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.MobileNumber) == "" {
		return fmt.Errorf("mobileNumber is required")
	}
	if strings.TrimSpace(req.IdentityType) == "" {
		return fmt.Errorf("identityType is required")
	}
	if strings.TrimSpace(req.IDNumber) == "" {
		return fmt.Errorf("idNumber is required")
	}
	if strings.TrimSpace(req.PhysicalAddress) == "" {
		return fmt.Errorf("physicalAddress is required")
	}
	if strings.TrimSpace(req.ThirdPartyVerificationURL) == "" {
		return fmt.Errorf("thirdPartyVerificationUrl is required")
	}
	return nil
}

func bankDepositDetails() map[string][]models.BankDepositDetail {
	return map[string][]models.BankDepositDetail{
		"ZAR": {
			{
				BankName:      "FNB",
				AccountName:   "Xago Holdings",
				AccountNumber: "62057334567",
				BranchCode:    "250145",
				SwiftBIC:      "FIRSZA22",
			},
		},
		"USD": {
			{
				BankName:      "Citibank",
				AccountName:   "Xago Inc",
				AccountNumber: "0123456789",
				BranchCode:    "021",
				SwiftBIC:      "CITIUS33",
			},
		},
	}
}

func generateDepositAddress() string {
	// Generate a mock XRP-like address
	rand.Seed(time.Now().UnixNano())
	return "r" + utils.GenerateToken()[:33]
}

func generateDepositTag() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(900000) + 100000
}
