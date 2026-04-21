package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mock/mockgatehub/internal/assets"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"go.uber.org/zap"
)

func (h *Handler) GetAccountConfirmation(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "walletAddress")
	if walletAddress == "" {
		h.sendError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	logger.Info("getting account confirmation statement", zap.String("wallet_address", walletAddress))

	h.sendStatementPDF(w, "account-confirmation.pdf")
}

func (h *Handler) GetAccountStatement(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "walletAddress")
	if walletAddress == "" {
		h.sendError(w, http.StatusBadRequest, "Wallet address is required")
		return
	}

	year := chi.URLParam(r, "year")
	if year == "" {
		h.sendError(w, http.StatusBadRequest, "Year is required")
		return
	}

	month := chi.URLParam(r, "month")
	if month == "" {
		h.sendError(w, http.StatusBadRequest, "Month is required")
		return
	}

	logger.Info("getting account statement",
		zap.String("wallet_address", walletAddress),
		zap.String("year", year),
		zap.String("month", month),
	)

	h.sendStatementPDF(w, "account-statement.pdf")
}

func (h *Handler) GetTransferConfirmation(w http.ResponseWriter, r *http.Request) {
	transactionUUID := chi.URLParam(r, "transactionUUID")
	if transactionUUID == "" {
		h.sendError(w, http.StatusBadRequest, "Transaction UUID is required")
		return
	}

	logger.Info("getting transfer confirmation statement", zap.String("transaction_uuid", transactionUUID))

	h.sendStatementPDF(w, "transfer-confirmation.pdf")
}

func (h *Handler) sendStatementPDF(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(assets.BlankPDF)
}
