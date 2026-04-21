package v1

import (
	"io"
	"net/http"
)

func (h *handlers) getAccountStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := h.users.UserForContext(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := h.gatehub.GetAccountStatement(ctx, wallet.ID)
	if err != nil {
		toHTTPError(w, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"account-statement.pdf\"")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, body)
}
