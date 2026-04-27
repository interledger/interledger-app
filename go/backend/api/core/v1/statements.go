package v1

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gitlab.com/fynbos/backend/api/apperrors"
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

	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, apperrors.ErrCodeBadRequest, "year must be a valid number")
		return
	}

	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, apperrors.ErrCodeBadRequest, "month must be a valid number")
		return
	}
	if month < 1 || month > 12 {
		apperrors.WriteAppError(w, r, http.StatusBadRequest, apperrors.ErrCodeBadRequest, "month must be between 1 and 12")
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
