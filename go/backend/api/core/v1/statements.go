package v1

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/api/apperrors"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func (h *handlers) getAccountStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.users.UserForContext(ctx); err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "year must be a valid number")
		return
	}

	month, err := strconv.Atoi(chi.URLParam(r, "month"))
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "month must be a valid number")
		return
	}
	if month < 1 || month > 12 {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "month must be between 1 and 12")
		return
	}

	body, err := h.gatehub.GetAccountStatement(ctx, wallet.ID, year, month)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"account-statement-%d-%02d.pdf\"", year, month))
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, body); err != nil {
		log.Warn("failed to stream account statement", zap.Error(err))
	}
}

func (h *handlers) getAccountConfirmation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.users.UserForContext(ctx); err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	body, err := h.gatehub.GetAccountConfirmation(ctx, wallet.ID)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\"account-confirmation.pdf\"")
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, body); err != nil {
		log.Warn("failed to stream account confirmation", zap.Error(err))
	}
}

func (h *handlers) getTransactionStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.users.UserForContext(ctx); err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}

	txID := chi.URLParam(r, "id")
	if txID == "" {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "transaction id is required")
		return
	}

	if err := uuid.Validate(txID); err != nil {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "invalid transaction id")
		return
	}

	body, err := h.gatehub.GetTransactionStatement(ctx, wallet.ID, txID)
	if err != nil {
		apperrors.ToHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"transaction-statement-%s.pdf\"", txID))
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, body); err != nil {
		log.Warn("failed to stream transaction statement", zap.Error(err))
	}
}
