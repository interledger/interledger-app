package handler

import (
	"net/http"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
)

// ListCurrencies handles GET /xago/v1/currencies
// Returns bank details in nested format (matches backend expectations).
// This format aligns with go/backend/providers/xago/external/types.go
func (h *Handler) ListCurrencies(w http.ResponseWriter, r *http.Request) {
	resp := []models.CurrencyNested{
		{
			CurrencyCode:    "ZAR",
			Name:            "South African Rand",
			Symbol:          "R",
			DepositEnabled:  true,
			WithdrawEnabled: true,
			MarketEnabled:   true,
			BankingProviders: []models.BankingProvider{
				{
					Name:             "FNB",
					DepositAvailable: true,
					DepositFields: models.DepositFields{
						BankName:       "FNB",
						AccountName:    "Xago Holdings",
						AccountNumber:  "62057334567",
						BranchCode:     "250145",
						BankAddress:    "FNB Building, Sandton, South Africa",
						AccountAddress: "Xago Holdings, Cape Town, South Africa",
						SwiftBIC:       "FIRSZA22",
					},
				},
			},
		},
		{
			CurrencyCode:    "USD",
			Name:            "US Dollar",
			Symbol:          "$",
			DepositEnabled:  true,
			WithdrawEnabled: true,
			MarketEnabled:   false,
			BankingProviders: []models.BankingProvider{
				{
					Name:             "Citibank",
					DepositAvailable: true,
					DepositFields: models.DepositFields{
						BankName:       "Citibank",
						AccountName:    "Xago Inc",
						AccountNumber:  "0123456789",
						BranchCode:     "021",
						BankAddress:    "Citibank Tower, New York, NY",
						AccountAddress: "Xago Inc, Delaware, USA",
						SwiftBIC:       "CITIUS33",
					},
				},
			},
		},
	}

	h.sendJSON(w, http.StatusOK, resp)
}
