package handler

import (
	"net/http"

	"gitlab.com/fynbos/mockxago/internal/models"
)

// ListCurrencies handles GET /xago/v1/currencies
// Returns hardcoded bank details for ZAR and USD.
func (h *Handler) ListCurrencies(w http.ResponseWriter, r *http.Request) {
	resp := []models.CurrencyResponse{
		{
			CurrencyID:    "ZAR",
			CurrencyName:  "South African Rand",
			BankName:      "FNB",
			AccountName:   "Xago Holdings",
			AccountNumber: "62057334567",
			BranchCode:    "250145",
			SwiftBIC:      "FIRSZA22",
		},
		{
			CurrencyID:    "USD",
			CurrencyName:  "US Dollar",
			BankName:      "Citibank",
			AccountName:   "Xago Inc",
			AccountNumber: "0123456789",
			BranchCode:    "021",
			SwiftBIC:      "CITIUS33",
		},
	}

	h.sendJSON(w, http.StatusOK, resp)
}
